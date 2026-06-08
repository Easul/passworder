package com.passworder

import android.annotation.SuppressLint
import android.content.ActivityNotFoundException
import android.content.ContentValues
import android.content.Intent
import android.net.Uri
import android.os.Build
import android.os.Bundle
import android.os.Environment
import android.provider.MediaStore
import android.util.Log
import android.view.View
import android.webkit.JavascriptInterface
import android.webkit.MimeTypeMap
import android.webkit.URLUtil
import android.webkit.ValueCallback
import android.webkit.WebChromeClient
import android.webkit.WebResourceError
import android.webkit.WebResourceRequest
import android.webkit.WebSettings
import android.webkit.WebView
import android.webkit.WebViewClient
import android.widget.Button
import android.widget.TextView
import android.widget.Toast
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import mobilebridge.Mobilebridge
import java.io.ByteArrayInputStream
import java.io.File
import java.io.IOException
import java.io.InputStream
import java.net.HttpURLConnection
import java.net.URL
import java.util.Base64
import java.util.concurrent.CountDownLatch
import kotlin.concurrent.thread

class MainActivity : AppCompatActivity() {

    companion object {
        private const val TAG = "Passworder"
        private const val SERVER_STARTUP_TIMEOUT_MS = 20_000L
        private const val SERVER_POLL_INTERVAL_MS = 400L
    }

    private lateinit var webView: WebView
    private lateinit var loadingContainer: View
    private lateinit var statusText: TextView
    private lateinit var retryButton: Button
    private var fileChooserCallback: ValueCallback<Array<Uri>>? = null
    private var serverStarted = false

    private val serverUrl: String
        get() = "http://${BuildConfig.LOCAL_SERVER_HOST}:${BuildConfig.LOCAL_SERVER_PORT}"

    private val localServerAuthority: String
        get() = "${BuildConfig.LOCAL_SERVER_HOST}:${BuildConfig.LOCAL_SERVER_PORT}"

    private val fileChooserLauncher = registerForActivityResult(ActivityResultContracts.StartActivityForResult()) { result ->
        val callback = fileChooserCallback ?: return@registerForActivityResult
        callback.onReceiveValue(WebChromeClient.FileChooserParams.parseResult(result.resultCode, result.data))
        fileChooserCallback = null
    }

    @SuppressLint("SetJavaScriptEnabled")
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        webView = findViewById(R.id.webview)
        loadingContainer = findViewById(R.id.loading_container)
        statusText = findViewById(R.id.status_text)
        retryButton = findViewById(R.id.retry_button)

        webView.settings.apply {
            javaScriptEnabled = true
            domStorageEnabled = true
            allowFileAccess = true
            allowContentAccess = true
            cacheMode = WebSettings.LOAD_DEFAULT
            mixedContentMode = WebSettings.MIXED_CONTENT_ALWAYS_ALLOW
            setSupportMultipleWindows(true)
        }
        webView.addJavascriptInterface(PassworderAndroidBridge(), "PassworderAndroid")
        webView.webChromeClient = createWebChromeClient()
        webView.webViewClient = createWebViewClient()
        webView.setDownloadListener { url, userAgent, contentDisposition, mimeType, _ ->
            handleDownloadRequest(url, userAgent, contentDisposition, mimeType)
        }

        retryButton.setOnClickListener { launchLocalServerAndLoad() }
        launchLocalServerAndLoad()
    }

    private fun createWebChromeClient(): WebChromeClient {
        return object : WebChromeClient() {
            override fun onShowFileChooser(
                webView: WebView?,
                filePathCallback: ValueCallback<Array<Uri>>?,
                fileChooserParams: FileChooserParams?
            ): Boolean {
                this@MainActivity.fileChooserCallback?.onReceiveValue(null)
                this@MainActivity.fileChooserCallback = filePathCallback
                val chooserIntent = try {
                    fileChooserParams?.createIntent() ?: buildFallbackChooserIntent()
                } catch (_: Exception) {
                    buildFallbackChooserIntent()
                }

                return try {
                    fileChooserLauncher.launch(chooserIntent)
                    true
                } catch (_: ActivityNotFoundException) {
                    this@MainActivity.fileChooserCallback = null
                    toast("当前设备没有可用的文件选择器")
                    false
                }
            }

            override fun onCreateWindow(view: WebView?, isDialog: Boolean, isUserGesture: Boolean, resultMsg: android.os.Message?): Boolean {
                val transport = resultMsg?.obj as? WebView.WebViewTransport ?: return false
                val transientWebView = WebView(this@MainActivity)
                transientWebView.webViewClient = object : WebViewClient() {
                    override fun shouldOverrideUrlLoading(view: WebView?, request: WebResourceRequest?): Boolean {
                        val uri = request?.url ?: return false
                        openExternalUri(uri)
                        return true
                    }
                }
                transport.webView = transientWebView
                resultMsg.sendToTarget()
                return true
            }
        }
    }

    private fun createWebViewClient(): WebViewClient {
        return object : WebViewClient() {
            override fun shouldOverrideUrlLoading(view: WebView?, request: WebResourceRequest?): Boolean {
                val uri = request?.url ?: return false
                return if (shouldOpenExternally(uri)) {
                    openExternalUri(uri)
                    true
                } else {
                    false
                }
            }

            override fun onReceivedError(view: WebView?, request: WebResourceRequest?, error: WebResourceError?) {
                if (request?.isForMainFrame == true) {
                    showError("页面加载失败，请重试")
                }
            }
        }
    }

    private fun buildFallbackChooserIntent(): Intent {
        return Intent(Intent.ACTION_GET_CONTENT).apply {
            addCategory(Intent.CATEGORY_OPENABLE)
            type = "*/*"
        }
    }

    private fun launchLocalServerAndLoad() {
        showLoading("正在启动本地服务…", false)
        thread(name = "passworder-bootstrap") {
            try {
                startEmbeddedServer()
                waitForServerReady()
                runOnUiThread {
                    webView.loadUrl(serverUrl)
                    loadingContainer.visibility = View.GONE
                    webView.visibility = View.VISIBLE
                }
            } catch (e: Exception) {
                Log.e(TAG, "Failed to start local server", e)
                showError(e.message ?: "本地服务启动失败")
            }
        }
    }

    private fun startEmbeddedServer() {
        if (serverStarted || isServerReachable()) {
            serverStarted = true
            return
        }
        Mobilebridge.touch()

        val appDataDir = File(filesDir, "passworder")
        val storageDir = File(appDataDir, "storage")
        if (!storageDir.exists() && !storageDir.mkdirs()) {
            throw IOException("无法创建存储目录")
        }
        val dbFile = File(appDataDir, "password.db")
        val error = Mobilebridge.startServer(
            BuildConfig.LOCAL_SERVER_HOST,
            BuildConfig.LOCAL_SERVER_PORT.toLong(),
            dbFile.absolutePath,
            storageDir.absolutePath,
        )
        if (!error.isNullOrBlank()) {
            throw IOException(error)
        }
        serverStarted = true
    }

    private fun waitForServerReady() {
        val deadline = System.currentTimeMillis() + SERVER_STARTUP_TIMEOUT_MS
        while (System.currentTimeMillis() < deadline) {
            if (isServerReachable()) {
                return
            }
            Thread.sleep(SERVER_POLL_INTERVAL_MS)
        }
        throw IOException("等待本地服务超时")
    }

    private fun isServerReachable(): Boolean {
        return try {
            val connection = URL(serverUrl).openConnection() as HttpURLConnection
            connection.requestMethod = "GET"
            connection.connectTimeout = 800
            connection.readTimeout = 800
            connection.instanceFollowRedirects = false
            connection.connect()
            val code = connection.responseCode
            connection.disconnect()
            code in 200..499
        } catch (_: Exception) {
            false
        }
    }

    private fun handleDownloadRequest(url: String, userAgent: String?, contentDisposition: String?, mimeType: String?) {
        thread(name = "passworder-download") {
            try {
                val connection = URL(url).openConnection() as HttpURLConnection
                connection.requestMethod = "GET"
                if (!userAgent.isNullOrBlank()) {
                    connection.setRequestProperty("User-Agent", userAgent)
                }
                val token = webView.evaluateJavascriptSync("localStorage.getItem('token')")
                if (!token.isNullOrBlank() && token != "null") {
                    connection.setRequestProperty("Authorization", token)
                }
                connection.connect()
                if (connection.responseCode >= 400) {
                    throw IOException("下载失败: HTTP ${connection.responseCode}")
                }

                val filename = URLUtil.guessFileName(url, contentDisposition, mimeType)
                connection.inputStream.use { input ->
                    saveStreamToDownloads(input, filename, connection.contentType ?: mimeType ?: "application/octet-stream")
                }
                connection.disconnect()
                toast("文件已保存到下载目录")
            } catch (e: Exception) {
                Log.e(TAG, "Download failed", e)
                toast("下载失败")
            }
        }
    }

    private fun saveStreamToDownloads(input: InputStream, filename: String, mimeType: String) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            val values = ContentValues().apply {
                put(MediaStore.Downloads.DISPLAY_NAME, filename)
                put(MediaStore.Downloads.MIME_TYPE, mimeType)
                put(MediaStore.Downloads.RELATIVE_PATH, Environment.DIRECTORY_DOWNLOADS)
            }
            val uri = contentResolver.insert(MediaStore.Downloads.EXTERNAL_CONTENT_URI, values)
                ?: throw IOException("无法创建下载文件")
            contentResolver.openOutputStream(uri)?.use { output ->
                input.copyTo(output)
            } ?: throw IOException("无法写入下载文件")
            return
        }

        val downloadsDir = getExternalFilesDir(Environment.DIRECTORY_DOWNLOADS)
            ?: throw IOException("无法访问下载目录")
        val target = File(downloadsDir, filename)
        target.outputStream().use { output ->
            input.copyTo(output)
        }
    }

    private fun shouldOpenExternally(uri: Uri): Boolean {
        val scheme = uri.scheme ?: return false
        if (scheme != "http" && scheme != "https") {
            return true
        }
        return uri.authority != localServerAuthority
    }

    private fun openExternalUri(uri: Uri) {
        try {
            startActivity(Intent(Intent.ACTION_VIEW, uri))
        } catch (e: Exception) {
            Log.e(TAG, "Failed to open external uri", e)
            toast("无法打开外部链接")
        }
    }

    private fun showLoading(message: String, showRetry: Boolean) {
        runOnUiThread {
            webView.visibility = View.GONE
            loadingContainer.visibility = View.VISIBLE
            statusText.text = message
            retryButton.visibility = if (showRetry) View.VISIBLE else View.GONE
        }
    }

    private fun showError(message: String) {
        showLoading(message, true)
    }

    private fun toast(message: String) {
        runOnUiThread {
            Toast.makeText(this, message, Toast.LENGTH_SHORT).show()
        }
    }

    private fun WebView.evaluateJavascriptSync(script: String): String? {
        return try {
            val latch = CountDownLatch(1)
            var result: String? = null
            runOnUiThread {
                evaluateJavascript(script) { value ->
                    result = value?.trim('"')
                    latch.countDown()
                }
            }
            latch.await()
            result
        } catch (e: Exception) {
            Log.e(TAG, "evaluateJavascriptSync failed", e)
            null
        }
    }

    inner class PassworderAndroidBridge {
        @JavascriptInterface
        fun saveBase64File(base64Data: String, filename: String, mimeType: String) {
            thread(name = "passworder-blob-save") {
                try {
                    val bytes = Base64.getDecoder().decode(base64Data)
                    ByteArrayInputStream(bytes).use { input ->
                        saveStreamToDownloads(input, filename, mimeType.ifBlank { guessMimeType(filename) })
                    }
                    toast("文件已保存到下载目录")
                } catch (e: Exception) {
                    Log.e(TAG, "saveBase64File failed", e)
                    toast("保存文件失败")
                }
            }
        }
    }

    private fun guessMimeType(filename: String): String {
        val extension = MimeTypeMap.getFileExtensionFromUrl(filename)
        return MimeTypeMap.getSingleton().getMimeTypeFromExtension(extension.lowercase())
            ?: "application/octet-stream"
    }

    override fun onBackPressed() {
        if (webView.canGoBack()) {
            webView.goBack()
        } else {
            super.onBackPressed()
        }
    }

    override fun onDestroy() {
        fileChooserCallback?.onReceiveValue(null)
        fileChooserCallback = null
        if (serverStarted) {
            Mobilebridge.stopServer()
            serverStarted = false
        }
        super.onDestroy()
    }
}

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
import android.provider.Settings
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
import androidx.core.content.FileProvider
import mobilebridge.Mobilebridge
import java.io.ByteArrayInputStream
import java.io.File
import java.io.IOException
import java.io.InputStream
import java.net.HttpURLConnection
import java.net.NetworkInterface
import java.net.URL
import java.util.Base64
import java.util.concurrent.CountDownLatch
import kotlin.concurrent.thread
import org.json.JSONArray
import org.json.JSONObject

class MainActivity : AppCompatActivity() {

    companion object {
        private const val TAG = "Passworder"
        private const val SERVER_STARTUP_TIMEOUT_MS = 20_000L
        private const val SERVER_POLL_INTERVAL_MS = 400L
        private const val PREF_SHOW_LAN_ACCESS_HINT = "show_lan_access_hint_after_load"
    }

    private lateinit var webView: WebView
    private lateinit var loadingContainer: View
    private lateinit var statusText: TextView
    private lateinit var retryButton: Button
    private var fileChooserCallback: ValueCallback<Array<Uri>>? = null
    private var serverStarted = false
    @Volatile private var showLanAccessHintAfterLoad = false

    private val serverUrl: String
        get() = "http://${BuildConfig.LOCAL_SERVER_WEB_HOST}:${BuildConfig.LOCAL_SERVER_PORT}"

    private val localServerAuthority: String
        get() = "${BuildConfig.LOCAL_SERVER_WEB_HOST}:${BuildConfig.LOCAL_SERVER_PORT}"

    private val fileChooserLauncher = registerForActivityResult(ActivityResultContracts.StartActivityForResult()) { result ->
        val callback = fileChooserCallback ?: return@registerForActivityResult
        callback.onReceiveValue(WebChromeClient.FileChooserParams.parseResult(result.resultCode, result.data))
        fileChooserCallback = null
    }

    private val previewFileLauncher = registerForActivityResult(ActivityResultContracts.StartActivityForResult()) {
        clearPreviewCache()
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
        showLanAccessHintAfterLoad = consumePendingLanAccessHint()
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
                    if (showLanAccessHintAfterLoad) {
                        showLanAccessHintAfterLoad = false
                        showLanAccessHint()
                    }
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
            BuildConfig.LOCAL_SERVER_LISTEN_HOST,
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

    private fun showLanAccessHint() {
        val urls = findLanAccessUrls()
        if (urls.isEmpty()) {
            toast("未找到可用于局域网访问的手机 IP")
            return
        }
        Toast.makeText(this, "请使用手机 IP 访问：\n${urls.joinToString("\n")}", Toast.LENGTH_LONG).show()
    }

    private fun findLanAccessUrls(): List<String> {
        val wifiUrls = mutableListOf<String>()
        val otherUrls = mutableListOf<String>()
        for (networkInterface in NetworkInterface.getNetworkInterfaces()) {
            if (!networkInterface.isUp || networkInterface.isLoopback) continue
            for (address in networkInterface.inetAddresses) {
                val hostAddress = address.hostAddress ?: continue
                if (address.isLoopbackAddress || hostAddress.contains(':')) continue
                val url = "http://$hostAddress:${BuildConfig.LOCAL_SERVER_PORT}"
                if (networkInterface.name.startsWith("wlan")) {
                    wifiUrls.add(url)
                } else {
                    otherUrls.add(url)
                }
            }
        }
        return wifiUrls + otherUrls
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

        @JavascriptInterface
        fun openBase64File(base64Data: String, filename: String, mimeType: String) {
            thread(name = "passworder-blob-open") {
                try {
                    val target = writePreviewFile(base64Data, filename)
                    val uri = FileProvider.getUriForFile(
                        this@MainActivity,
                        "${BuildConfig.APPLICATION_ID}.fileprovider",
                        target,
                    )
                    val intent = Intent(Intent.ACTION_VIEW).apply {
                        setDataAndType(uri, mimeType.ifBlank { guessMimeType(filename) })
                        addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
                    }
                    runOnUiThread {
                        previewFileLauncher.launch(Intent.createChooser(intent, "打开文件"))
                    }
                } catch (e: ActivityNotFoundException) {
                    clearPreviewCache()
                    Log.e(TAG, "No viewer for file", e)
                    toast("当前设备没有可用的文件查看器")
                } catch (e: Exception) {
                    clearPreviewCache()
                    Log.e(TAG, "openBase64File failed", e)
                    toast("打开文件失败")
                }
            }
        }

        @JavascriptInterface
        fun restartServer(showLanAccessHint: Boolean) {
            getPreferences(MODE_PRIVATE)
                .edit()
                .putBoolean(PREF_SHOW_LAN_ACCESS_HINT, showLanAccessHint)
                .apply()
            runOnUiThread {
                webView.stopLoading()
                restartApp()
            }
        }

        @JavascriptInterface
        fun translate(requestId: String, baseUrl: String, apiKey: String, model: String, text: String) {
            thread(name = "passworder-translate") {
                try {
                    val translated = requestTranslation(baseUrl, apiKey, model, text)
                    sendTranslatorResult(requestId, translated, null)
                } catch (e: Exception) {
                    Log.e(TAG, "translate failed", e)
                    sendTranslatorResult(requestId, null, e.message ?: "翻译失败")
                }
            }
        }

        @JavascriptInterface
        fun showTranslatorOverlay(baseUrl: String, apiKey: String, model: String) {
            runOnUiThread {
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M && !Settings.canDrawOverlays(this@MainActivity)) {
                    toast("请开启悬浮窗权限后再次点击翻译")
                    val intent = Intent(
                        Settings.ACTION_MANAGE_OVERLAY_PERMISSION,
                        Uri.parse("package:$packageName"),
                    )
                    startActivity(intent)
                    return@runOnUiThread
                }
                TranslatorOverlayService.start(this@MainActivity, baseUrl, apiKey, model)
                toast("翻译悬浮窗已开启")
                moveTaskToBack(true)
            }
        }
    }

    private fun requestTranslation(baseUrl: String, apiKey: String, model: String, text: String): String {
        val endpoint = translationEndpoint(baseUrl)
        val targetLanguage = if (Regex("[\\u4e00-\\u9fff]").containsMatchIn(text)) "English" else "Chinese"
        val payload = JSONObject()
            .put("model", model)
            .put("stream", false)
            .put(
                "messages",
                JSONArray().put(
                    JSONObject()
                        .put("role", "user")
                        .put("content", "Translate the following text into $targetLanguage. Return only the translation, with no explanation.\n\n$text")
                )
            )
            .toString()

        val connection = URL(endpoint).openConnection() as HttpURLConnection
        try {
            connection.requestMethod = "POST"
            connection.connectTimeout = 20_000
            connection.readTimeout = 60_000
            connection.doOutput = true
            connection.setRequestProperty("Content-Type", "application/json")
            connection.setRequestProperty("Authorization", "Bearer $apiKey")
            connection.outputStream.use { output -> output.write(payload.toByteArray(Charsets.UTF_8)) }
            val responseText = if (connection.responseCode in 200..299) {
                connection.inputStream.bufferedReader().use { it.readText() }
            } else {
                val errorText = connection.errorStream?.bufferedReader()?.use { it.readText() }.orEmpty()
                throw IOException(parseTranslationError(errorText, connection.responseCode))
            }
            return JSONObject(responseText)
                .getJSONArray("choices")
                .getJSONObject(0)
                .getJSONObject("message")
                .getString("content")
                .trim()
        } finally {
            connection.disconnect()
        }
    }

    private fun translationEndpoint(baseUrl: String): String {
        val normalized = baseUrl.trim().trimEnd('/')
        if (normalized.endsWith("/chat/completions") || normalized.endsWith("/completions")) {
            return normalized
        }
        return "$normalized/chat/completions"
    }

    private fun parseTranslationError(errorText: String, code: Int): String {
        return try {
            JSONObject(errorText).optJSONObject("error")?.optString("message")?.takeIf { it.isNotBlank() }
                ?: "翻译失败：HTTP $code"
        } catch (_: Exception) {
            "翻译失败：HTTP $code"
        }
    }

    private fun sendTranslatorResult(requestId: String, content: String?, error: String?) {
        val script = "window.Passworder?.receiveTranslationResult(${JSONObject.quote(requestId)}, ${JSONObject.quote(content)}, ${JSONObject.quote(error)})"
        runOnUiThread { webView.evaluateJavascript(script, null) }
    }

    private fun consumePendingLanAccessHint(): Boolean {
        val preferences = getPreferences(MODE_PRIVATE)
        val shouldShow = preferences.getBoolean(PREF_SHOW_LAN_ACCESS_HINT, false)
        if (shouldShow) {
            preferences.edit().remove(PREF_SHOW_LAN_ACCESS_HINT).apply()
        }
        return shouldShow
    }

    private fun restartApp() {
        val intent = packageManager.getLaunchIntentForPackage(packageName) ?: return
        intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK)
        startActivity(intent)
        finish()
        Runtime.getRuntime().exit(0)
    }

    private fun writePreviewFile(base64Data: String, filename: String): File {
        val previewDir = File(cacheDir, "preview")
        if (!previewDir.exists() && !previewDir.mkdirs()) {
            throw IOException("无法创建预览目录")
        }
        previewDir.listFiles()?.forEach { it.delete() }
        val safeFilename = sanitizeFilename(filename.ifBlank { "preview" })
        val target = File(previewDir, safeFilename)
        val bytes = Base64.getDecoder().decode(base64Data)
        target.outputStream().use { output -> output.write(bytes) }
        return target
    }

    private fun clearPreviewCache() {
        File(cacheDir, "preview").deleteRecursively()
    }

    private fun sanitizeFilename(filename: String): String {
        return filename.replace(Regex("[\\\\/:*?\"<>|]"), "_")
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
        clearPreviewCache()
        if (serverStarted) {
            Mobilebridge.stopServer()
            serverStarted = false
        }
        super.onDestroy()
    }
}

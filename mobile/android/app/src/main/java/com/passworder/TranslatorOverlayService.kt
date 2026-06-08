package com.passworder

import android.app.Service
import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import android.content.Intent
import android.graphics.Color
import android.graphics.Typeface
import android.graphics.drawable.GradientDrawable
import android.os.Build
import android.os.Handler
import android.os.IBinder
import android.os.Looper
import android.view.Gravity
import android.view.MotionEvent
import android.view.View
import android.view.WindowManager
import android.view.inputmethod.InputMethodManager
import android.widget.Button
import android.widget.EditText
import android.widget.LinearLayout
import android.widget.TextView
import android.widget.Toast
import java.io.IOException
import java.net.HttpURLConnection
import java.net.URL
import kotlin.concurrent.thread
import org.json.JSONArray
import org.json.JSONObject

class TranslatorOverlayService : Service() {
    companion object {
        private const val EXTRA_BASE_URL = "base_url"
        private const val EXTRA_API_KEY = "api_key"
        private const val EXTRA_MODEL = "model"

        fun start(context: Context, baseUrl: String, apiKey: String, model: String) {
            val intent = Intent(context, TranslatorOverlayService::class.java)
                .putExtra(EXTRA_BASE_URL, baseUrl)
                .putExtra(EXTRA_API_KEY, apiKey)
                .putExtra(EXTRA_MODEL, model)
            context.startService(intent)
        }
    }

    private val mainHandler = Handler(Looper.getMainLooper())
    private lateinit var windowManager: WindowManager
    private var overlayView: View? = null
    private var baseUrl = ""
    private var apiKey = ""
    private var model = ""
    private lateinit var inputView: EditText
    private lateinit var outputView: TextView
    private lateinit var translateButton: Button
    private var overlayParams: WindowManager.LayoutParams? = null
    private var isCollapsed = false

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        baseUrl = intent?.getStringExtra(EXTRA_BASE_URL).orEmpty()
        apiKey = intent?.getStringExtra(EXTRA_API_KEY).orEmpty()
        model = intent?.getStringExtra(EXTRA_MODEL).orEmpty()
        if (overlayView == null) {
            showOverlay()
        } else {
            overlayView?.bringToFront()
        }
        return START_STICKY
    }

    private fun showOverlay() {
        windowManager = getSystemService(WINDOW_SERVICE) as WindowManager
        val params = WindowManager.LayoutParams(
            dp(280),
            WindowManager.LayoutParams.WRAP_CONTENT,
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY
            } else {
                @Suppress("DEPRECATION")
                WindowManager.LayoutParams.TYPE_PHONE
            },
            WindowManager.LayoutParams.FLAG_LAYOUT_IN_SCREEN,
            android.graphics.PixelFormat.TRANSLUCENT,
        ).apply {
            gravity = Gravity.TOP or Gravity.START
            x = dp(24)
            y = dp(96)
            softInputMode = WindowManager.LayoutParams.SOFT_INPUT_ADJUST_RESIZE
        }
        overlayParams = params
        val root = buildOverlayView(params)
        overlayView = root
        windowManager.addView(root, params)
        inputView.requestFocus()
        (getSystemService(INPUT_METHOD_SERVICE) as InputMethodManager).showSoftInput(inputView, InputMethodManager.SHOW_IMPLICIT)
    }

    private fun buildOverlayView(params: WindowManager.LayoutParams): View {
        val root = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(dp(10), dp(8), dp(10), dp(10))
            background = GradientDrawable().apply {
                setColor(Color.WHITE)
                cornerRadius = dp(12).toFloat()
                setStroke(1, Color.rgb(226, 232, 240))
            }
            elevation = dp(8).toFloat()
            setOnTouchListener(dragListener(params))
        }

        val header = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setOnTouchListener(dragListener(params))
        }
        header.addView(TextView(this).apply {
            text = "🌐 翻译"
            textSize = 15f
            setTextColor(Color.rgb(15, 23, 42))
            setTypeface(typeface, Typeface.BOLD)
        }, LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f))
        header.addView(Button(this).apply {
            text = "●"
            setOnClickListener { collapseOverlay() }
        }, LinearLayout.LayoutParams(dp(40), dp(36)))
        header.addView(Button(this).apply {
            text = "×"
            setOnClickListener { stopSelf() }
        }, LinearLayout.LayoutParams(dp(40), dp(36)))
        root.addView(header)

        inputView = EditText(this).apply {
            hint = "输入中文或英文"
            textSize = 14f
            minLines = 2
            maxLines = 4
            setSingleLine(false)
        }
        root.addView(inputView, LinearLayout.LayoutParams(LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT))

        val actions = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.END
        }
        actions.addView(Button(this).apply {
            text = "清空"
            setOnClickListener {
                inputView.setText("")
                outputView.text = ""
            }
        })
        translateButton = Button(this).apply {
            text = "翻译"
            setOnClickListener { translateCurrentInput() }
        }
        actions.addView(translateButton)
        root.addView(actions)

        outputView = TextView(this).apply {
            textSize = 14f
            setTextColor(Color.rgb(15, 23, 42))
            setTextIsSelectable(true)
            setPadding(dp(8), dp(8), dp(8), dp(8))
            background = GradientDrawable().apply {
                setColor(Color.rgb(248, 250, 252))
                cornerRadius = dp(8).toFloat()
                setStroke(1, Color.rgb(226, 232, 240))
            }
        }
        root.addView(outputView, LinearLayout.LayoutParams(LinearLayout.LayoutParams.MATCH_PARENT, dp(72)).apply {
            topMargin = dp(6)
            bottomMargin = dp(6)
        })

        root.addView(Button(this).apply {
            text = "复制结果"
            setOnClickListener { copyResult() }
        }, LinearLayout.LayoutParams(LinearLayout.LayoutParams.MATCH_PARENT, dp(40)))
        return root
    }

    private fun collapseOverlay() {
        val params = overlayParams ?: return
        val oldView = overlayView ?: return
        windowManager.removeView(oldView)
        isCollapsed = true
        params.width = dp(48)
        params.height = dp(48)
        val dot = buildCollapsedView(params)
        overlayView = dot
        windowManager.addView(dot, params)
    }

    private fun expandOverlay() {
        val params = overlayParams ?: return
        val oldView = overlayView ?: return
        windowManager.removeView(oldView)
        isCollapsed = false
        params.width = dp(280)
        params.height = WindowManager.LayoutParams.WRAP_CONTENT
        val root = buildOverlayView(params)
        overlayView = root
        windowManager.addView(root, params)
    }

    private fun buildCollapsedView(params: WindowManager.LayoutParams): View {
        return TextView(this).apply {
            text = "🌐"
            textSize = 22f
            gravity = Gravity.CENTER
            background = GradientDrawable().apply {
                shape = GradientDrawable.OVAL
                setColor(Color.rgb(79, 70, 229))
            }
            elevation = dp(8).toFloat()
            setOnTouchListener(dragListener(params) { expandOverlay() })
        }
    }

    private fun dragListener(params: WindowManager.LayoutParams, onTap: (() -> Unit)? = null): View.OnTouchListener {
        var startX = 0
        var startY = 0
        var downRawX = 0f
        var downRawY = 0f
        var moved = false
        return View.OnTouchListener { _, event ->
            when (event.action) {
                MotionEvent.ACTION_DOWN -> {
                    startX = params.x
                    startY = params.y
                    downRawX = event.rawX
                    downRawY = event.rawY
                    moved = false
                    true
                }
                MotionEvent.ACTION_MOVE -> {
                    val deltaX = event.rawX - downRawX
                    val deltaY = event.rawY - downRawY
                    moved = moved || kotlin.math.abs(deltaX) > dp(4) || kotlin.math.abs(deltaY) > dp(4)
                    params.x = startX + deltaX.toInt()
                    params.y = startY + deltaY.toInt()
                    overlayView?.let { windowManager.updateViewLayout(it, params) }
                    true
                }
                MotionEvent.ACTION_UP -> {
                    if (!moved) onTap?.invoke()
                    true
                }
                else -> false
            }
        }
    }

    private fun translateCurrentInput() {
        val text = inputView.text.toString().trim()
        if (text.isBlank()) {
            toast("请输入要翻译的内容")
            return
        }
        outputView.text = "翻译中..."
        translateButton.isEnabled = false
        thread(name = "passworder-overlay-translate") {
            try {
                val translated = requestTranslation(text)
                mainHandler.post { outputView.text = translated }
            } catch (e: Exception) {
                mainHandler.post {
                    outputView.text = ""
                    toast(e.message ?: "翻译失败")
                }
            } finally {
                mainHandler.post { translateButton.isEnabled = true }
            }
        }
    }

    private fun requestTranslation(text: String): String {
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
        val connection = URL(translationEndpoint(baseUrl)).openConnection() as HttpURLConnection
        try {
            connection.requestMethod = "POST"
            connection.connectTimeout = 20_000
            connection.readTimeout = 60_000
            connection.doOutput = true
            connection.setRequestProperty("Content-Type", "application/json")
            connection.setRequestProperty("Authorization", "Bearer $apiKey")
            connection.outputStream.use { it.write(payload.toByteArray(Charsets.UTF_8)) }
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

    private fun copyResult() {
        val result = outputView.text.toString().trim()
        if (result.isBlank() || result == "翻译中...") {
            toast("没有可复制的翻译结果")
            return
        }
        val clipboard = getSystemService(CLIPBOARD_SERVICE) as ClipboardManager
        clipboard.setPrimaryClip(ClipData.newPlainText("translation", result))
        toast("翻译结果已复制")
    }

    private fun toast(message: String) {
        mainHandler.post { Toast.makeText(this, message, Toast.LENGTH_SHORT).show() }
    }

    private fun dp(value: Int): Int = (value * resources.displayMetrics.density).toInt()

    override fun onDestroy() {
        overlayView?.let { windowManager.removeView(it) }
        overlayView = null
        super.onDestroy()
    }
}

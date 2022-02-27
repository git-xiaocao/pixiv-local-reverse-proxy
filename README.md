# Pixiv本地反向代理

用于WebView登录Pixiv

# 如何使用?

# Test

直接运行[main_test.go](./main_test.go)里的函数

![img.png](img.png)

Go版本1.17.1

## Android

`获取gomobile`

```console
go install golang.org/x/mobile/cmd/gomobile
```

配置环境变量(示例)

1. 指定`ANDROID_HOME`为`Android SDK`安装目录
2. 指定`ANDROID_NDK_HOME`为 `%ANDROID_HOME%\ndk\23.0.7599858`


`gomobile` 编译 参考[build_android](./build_android)

Gradle Groovy

```groovy
implementation files("pixiv_local_reverse_proxy.aar")
```

Gradle KTS

```kotlin
implementation(files("pixiv_local_reverse_proxy.aar"))
```

`WebView`设置代理
```kotlin
if (WebViewFeature.isFeatureSupported(WebViewFeature.PROXY_OVERRIDE)) {
    val proxyConfig: ProxyConfig = ProxyConfig.Builder()
        .addProxyRule("127.0.0.1:12345")
        .addDirect()
        .build()
    ProxyController.getInstance().setProxyOverride(
        proxyConfig,
        { command -> command?.run() },
    ) {
        //
    }
}
```

启动
```kotlin
PixivLocalReverseProxy.startServer(12345)
```

停止
```kotlin
PixivLocalReverseProxy.stopServer()
```

## Windows

`CGO`编译 参考[build_windows](./build_windows) 

`x64` 编译器设置
```console
go env -w GOARCH=amd64
go env -w CGO_ENABLED=0
```

`x86` 编译器设置
```console
go env -w GOARCH=386
go env -w CGO_ENABLED=1
```

C#声明
```cs
[DllImport(@"PixivLocalReverseProxy.dll")]
public static extern void StartServer(ushort bindPort);

[DllImport(@"PixivLocalReverseProxy.dll")]
public static extern void StopServer();
```

`Microsoft Edge WebView2`设置代理
```cs
Environment.SetEnvironmentVariable("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", $"--proxy-server=127.0.0.1:12345 --ignore-certificate-errors");
```

如果要获取登录成功后的`Uri`里的信息(也就是pixiv://account) 

需要使用`webView.CoreWebView2.NavigationStarting` 而不是`webView.NavigationStarting` 

见[MicrosoftEdge/WebView2Feedback/issues/2102](https://github.com/MicrosoftEdge/WebView2Feedback/issues/2102)

**注意系统代理比 `Environment.SetEnvironmentVariable`的优先级高**

### 感谢[dylech30th](https://github.com/dylech30th) 为我提供的C#语言 .NET平台 WinUI3 等基础知识讲解(恶补)

# iOS
`gomobile` 编译 参考[build_ios](./build_ios)

自己想办法 我的GTX750Ti显卡装不了黑苹果 没弄过iOS
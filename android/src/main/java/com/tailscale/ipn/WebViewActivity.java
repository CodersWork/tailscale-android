package com.tailscale.ipn;

import android.app.Activity;
import android.os.Bundle;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.webkit.WebSettings;
import android.webkit.WebResourceRequest;

import android.util.Log;
import android.net.Uri;
import android.content.Intent;

import java.net.URI;

import com.android.dingtalk.openauth.AuthLoginParam;
import com.android.dingtalk.openauth.IDDAuthApi;
import com.android.dingtalk.openauth.DDAuthApiFactory;

import com.tencent.mm.opensdk.openapi.IWXAPI;
import com.tencent.mm.opensdk.openapi.WXAPIFactory;
import com.tencent.mm.opensdk.modelmsg.SendAuth;

public final class WebViewActivity extends Activity {

    public static final String EXTRA_URL = "url";
    public static final String EXTRA_AUTHCODE = "authCode";
    public static final String EXTRA_STATE = "state";
    public static final String EXTRA_SOURCE = "source";
    public static final String ACTION_OPEN = "com.tailscale.ipn.ACTION_OPEN";
    public static final String ACTION_AUTH = "com.tailscale.ipn.ACTION_AUTH";
    public static final String ACTION_CLOSE = "com.tailscale.ipn.ACTION_CLOSE";

    private static final String APP_ID = "wxaae48dbe6f59d6e1";
    private IWXAPI api;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        Intent intent = getIntent();
        String action = intent.getAction();

        if (ACTION_OPEN.equals(action)) {
            openWebView();
        } else if (ACTION_CLOSE.equals(action)) {
            finish();
        } else if (ACTION_AUTH.equals(action)) {
            Log.d("Dingtalk", "called in WebViewActivity class: " + action);
            String authCode = intent.getStringExtra(EXTRA_AUTHCODE);
            String state = intent.getStringExtra(EXTRA_STATE);
            if (authCode != null && state != null) {
                WebView webView = new WebView(this);
                WebSettings webSettings = webView.getSettings();
                webSettings.setJavaScriptEnabled(true);
                setContentView(webView);
                setTitle("蜃境网络");
                webView.setWebViewClient(new WebViewClient() {
                    @Override
                    public boolean shouldOverrideUrlLoading(WebView view, WebResourceRequest request) {
                        String url_prefix = "https://sdp.full-mesh.cn/admin";
                        Uri uri = request.getUrl();
                        if (uri.toString().startsWith(url_prefix)) {
                            Log.d("Dingtalk", "shouldOverrideUrlLoading: " + uri.toString());
                            finish();
                            return true;
                        }
                        return false;
                    }
                });
                String source = intent.getStringExtra(EXTRA_SOURCE);
                if (source != null && source.equals("wechat_app")) {
                    String url = "https://sdp.full-mesh.cn/issuer/callback?wechat_app&code=" + authCode + "&state="
                            + state;
                    Log.d("Wechat", "WebViewActivity url: " + url);
                    webView.loadUrl(url);
                } else {
                    String url = "https://sdp.full-mesh.cn/issuer/callback?authCode=" + authCode + "&state=" + state;
                    Log.d("Dingtalk", "WebViewActivity url: " + url);
                    webView.loadUrl(url);
                }
            }
        }
    }

    private void openWebView() {
        WebView webView = new WebView(this);
        WebSettings webSettings = webView.getSettings();
        webSettings.setJavaScriptEnabled(true);
        setContentView(webView);
        setTitle("蜃境网络");

        String url = getIntent().getStringExtra(EXTRA_URL);
        Log.i("Dingtalk", "called in WebViewActivity class: " + url);
        if (url != null) {
            webView.setWebViewClient(new WebViewClient() {
                @Override
                public boolean shouldOverrideUrlLoading(WebView view, WebResourceRequest request) {
                    String dingtalk_prefix = "https://login.dingtalk.com/oauth2/auth";
                    String wechat_prefix = "https://open.weixin.qq.com/connect/qrconnect";
                    Uri uri = request.getUrl();
                    if (uri.toString().startsWith(dingtalk_prefix)) {
                        Log.d("Dingtalk", "shouldOverrideUrlLoading: " + uri.toString());
                        loginDingtalk(uri.getQueryParameter("state"));
                        finish();
                        return true;
                    }
                    if (uri.toString().startsWith(wechat_prefix)) {
                        Log.d("Wechat", "shouldOverrideUrlLoading: " + uri.toString());
                        loginWechat(uri.getQueryParameter("state"));
                        finish();
                        return true;
                    }
                    return false;
                }
            });

            webView.loadUrl(url);
        }
    }

    private void loginDingtalk(String state) {
        AuthLoginParam.AuthLoginParamBuilder builder = AuthLoginParam.AuthLoginParamBuilder.newBuilder();
        builder.appId("dingmaup5nlhpi7ixgqz");
        builder.redirectUri("https://sdp.full-mesh.cn/issuer/callback");
        builder.responseType("code");
        builder.scope("openid");
        builder.nonce("myNonce");
        builder.state(state);
        builder.prompt("consent");
        IDDAuthApi authApi = DDAuthApiFactory.createDDAuthApi(getApplicationContext(), builder.build());
        authApi.authLogin();
    }

    private void loginWechat(String state) {
        api = WXAPIFactory.createWXAPI(this, APP_ID, true);

        api.registerApp(APP_ID);

        final SendAuth.Req req = new SendAuth.Req();
        req.scope = "snsapi_userinfo"; // 只能填 snsapi_userinfo
        req.state = state;
        Log.d("Wechat", "sendReq state: " + state);
        api.sendReq(req);
    }
}

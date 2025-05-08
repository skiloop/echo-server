# 跳过 TLS 证书校验
Add-Type @"
using System.Net;
using System.Net.Security;
using System.Security.Cryptography.X509Certificates;

public class TrustAllCertsPolicy {
    public static void Ignore() {
        ServicePointManager.ServerCertificateValidationCallback =
            delegate (object sender, X509Certificate certificate, X509Chain chain, SslPolicyErrors sslPolicyErrors) { return true; };
    }
}
"@

[TrustAllCertsPolicy]::Ignore()

# 使用 Invoke-WebRequest 或 Invoke-RestMethod
$response = Invoke-WebRequest -Uri "https://localhost:9013/ja3" -Method Get
Write-Output "status code:"
Write-Output $response.StatusCode
Write-Output "headers:"
Write-Output $response.Headers
Write-Output "content:"
Write-Output $response.Content
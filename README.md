# F1WebViewer-SelfHosted
Self-hosted reverse-proxy for F1 web viewer, includes a web server at port 13331.

### Running Locally

Download binary from https://github.com/iebb/F1WebViewer-SelfHosted/releases, or build your own, or just `go run main.go`

### Running on a Server (Requires SSL to make DRM work):
Encrypted Media requires secure context to work, which means https is required except localhost. 

If you want to run it on your own server, you have to connect using HTTPS if you want to watch LIVE content. 

This tool provides some easy-to-use SSL options, or you can still manage yourself.

**Here are some scenarios:**
* If you have a public domain name pointed to your server, and have port 80 open
  * Run with `-ssl` parameter. 
  * It would automatically fetch a certificate through Lets Encrypt.

* If your server has a public domain name in cloudflare
  * Run with `-port 80`, or other cloudflare-enabled ports, and use their Flexible SSL.

* If you don't have a public domain name, or it's a LAN server, or something else
  * Run with `-self-sign` parameter. 
  * It would issue and use a self-signed certificate.
  * There would be a browser warning, you know what to do :D

* If you have some other web servers which is SSL-enabled and capable of reverse proxying
  * Do a reverse proxy. Make sure double slashes are preserved during this.

* Running with `-ssl` or `-self-sign` requires you opening the browser manually, as these commands are supposed to run in a server rather than your PC.

### Tutorial:

![image](https://user-images.githubusercontent.com/2127498/162486955-ca58805d-da15-43e0-9b4a-a54a31401a10.png)

Download binary from https://github.com/iebb/F1WebViewer-SelfHosted/releases, or build your own, or just `go run main.go`

You can also run this proxy in your server. That might be useful if your home country doesn't have F1TV.

### Extras:

Currently authentication and Reese84 Requests are routed through the public website as the solution might be immature.

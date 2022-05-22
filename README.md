# F1WebViewer-SelfHosted

Self-hosted reverse-proxy for F1 web viewer and includes a web server at port 13331. You can also run this proxy on a server if your home country doesn't have F1TV.

## Installation and Usage

Currently authentication and Reese84 Requests are routed through the public website as the solution might be immature.

If you are experiencing lag / stutter or other issues locally, try using Bitmovin Player in the Setting Tab. License is NOT required for local deploys.

### Compiled Binary

Download the latest binary [from the Releases](https://github.com/iebb/F1WebViewer-SelfHosted/releases) and run. You may need to run `chmod +x <filename>` to make the file executable on Mac / Linux

### Compile / Run from Source

Download and install Go, then

```bash
git clone https://github.com/iebb/F1WebViewer-SelfHosted.git
cd F1WebViewer-SelfHosted
go run . 
#or go build .
```

### Running on a Server (Requires SSL to make DRM work)

Streams with DRM (Encrypted Media) require secure context to work. If you want to run this on a server (aka not on localhost), you have to connect using HTTPS to watch LIVE content.

This tool provides some easy-to-use SSL options, or you can still manage yourself.

## Custom Arguments

| Command             | Notes
 :------------------ | :---------
| `-ssl`  | SSL support via Lets Encrypt. Requires a public IP and port 80 being open to fetch the certificate
| `-listen localhost:80` | Use a custom `host:port` instead of the default `localhost:13331`
| `-self-sign` | Issue and use a self-signed SSL certificate. Will display a warning in browsers

Running with `-ssl` or `-self-sign` requires you opening the browser manually, as these commands are supposed to run in a server rather than your PC.

## Screenshots

![image](https://user-images.githubusercontent.com/2127498/162486955-ca58805d-da15-43e0-9b4a-a54a31401a10.png)

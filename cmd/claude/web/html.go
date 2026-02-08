package web

const indexHTML = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
  <title>tofu claude</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@xterm/xterm@6.0.0/css/xterm.min.css">
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    html, body { height: 100%; background: #000; overflow: hidden; }
    #terminal { height: 100%; width: 100%; }
    #status {
      position: fixed; top: 8px; right: 8px;
      padding: 4px 10px; border-radius: 4px;
      font-family: monospace; font-size: 12px;
      color: #fff; background: rgba(80,80,80,0.8);
      z-index: 10; transition: opacity 0.3s;
    }
    #status.connected { background: rgba(0,120,0,0.8); }
    #status.disconnected { background: rgba(180,0,0,0.8); }
  </style>
</head>
<body>
  <div id="status">connecting...</div>
  <div id="terminal"></div>

  <script src="https://cdn.jsdelivr.net/npm/@xterm/xterm@6.0.0/lib/xterm.min.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/@xterm/addon-fit@0.11.0/lib/addon-fit.min.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/@xterm/addon-web-links@0.12.0/lib/addon-web-links.min.js"></script>
  <script>
    const term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: '"Cascadia Code", "Fira Code", "SF Mono", Menlo, monospace',
      theme: {
        background: '#1a1a2e',
        foreground: '#e0e0e0',
        cursor: '#e0e0e0',
        selectionBackground: 'rgba(255,255,255,0.2)',
      },
      allowProposedApi: true,
    });

    const fitAddon = new FitAddon.FitAddon();
    const webLinksAddon = new WebLinksAddon.WebLinksAddon();
    term.loadAddon(fitAddon);
    term.loadAddon(webLinksAddon);
    term.open(document.getElementById('terminal'));
    fitAddon.fit();

    const status = document.getElementById('status');
    let ws;
    let reconnectTimer;

    function connect() {
      const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
      ws = new WebSocket(proto + '//' + location.host + '/ws');
      ws.binaryType = 'arraybuffer';

      ws.onopen = () => {
        status.textContent = 'connected';
        status.className = 'connected';
        // Fade out status after 2s
        setTimeout(() => { status.style.opacity = '0'; }, 2000);
        // Send initial size
        sendResize();
      };

      ws.onmessage = (e) => {
        if (e.data instanceof ArrayBuffer) {
          term.write(new Uint8Array(e.data));
        } else {
          term.write(e.data);
        }
      };

      ws.onclose = () => {
        status.textContent = 'disconnected - reconnecting...';
        status.className = 'disconnected';
        status.style.opacity = '1';
        reconnectTimer = setTimeout(connect, 2000);
      };

      ws.onerror = () => {
        ws.close();
      };
    }

    function sendResize() {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'resize',
          cols: term.cols,
          rows: term.rows,
        }));
      }
    }

    // Terminal input -> WebSocket
    term.onData((data) => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(new TextEncoder().encode(data));
      }
    });

    // Handle resize
    window.addEventListener('resize', () => {
      fitAddon.fit();
      sendResize();
    });

    // Also send resize when fit changes dimensions
    term.onResize(() => {
      sendResize();
    });

    connect();
  </script>
</body>
</html>`

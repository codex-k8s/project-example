export type WsEnvelope = { type: string; payload: any };

export function connectWs(onMessage: (msg: WsEnvelope) => void): WebSocket {
  const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
  const ws = new WebSocket(`${proto}://${window.location.host}/ws`);
  ws.onmessage = (ev) => {
    try {
      const msg = JSON.parse(ev.data);
      onMessage(msg);
    } catch {
      // ignore
    }
  };
  return ws;
}


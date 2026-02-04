export type WsEnvelope = { type: string; payload: any };

export function connectWs(onMessage: (msg: WsEnvelope) => void): WebSocket {
  // Клиенты ходят только по https -> wss.
  const ws = new WebSocket(`wss://${window.location.host}/ws`);
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

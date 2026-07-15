import ws from 'k6/ws';
import { check } from 'k6';
import { Trend } from 'k6/metrics';

const msgLatency = new Trend('msg_latency', true);

export let options = {
  vus: 10,          // reduce to 10 clients
  duration: '20s',
};

export default function () {
  const url = 'ws://localhost:8080/ws';
  const myId = `vu-${__VU}-${Date.now()}`;

  const res = ws.connect(url, {}, function (socket) {
    socket.on('open', function () {
      socket.setInterval(function () {
        const payload = JSON.stringify({ id: myId, ts: Date.now() });
        socket.send(payload);
      }, 1000); // once per second
    });

    socket.on('message', function (data) {
      try {
        const msg = JSON.parse(data);
        if (msg.id === myId && msg.ts) {
          const latency = Date.now() - msg.ts;
          msgLatency.add(latency);
          check(latency, { 'latency under 20ms': (l) => l < 20 });
        }
      } catch (_) {}
    });

    socket.setTimeout(function () {
      socket.close();
    }, 18000);
  });

  check(res, { 'status is 101': (r) => r && r.status === 101 });
}
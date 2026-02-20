let pc = null;
let socket = null;

async function start() {
  const localVideo = document.getElementById("localVideo");
  const remoteVideo = document.getElementById("remoteVideo");

  const stream = await navigator.mediaDevices.getUserMedia({
    video: true,
    audio: false,
  });
  localVideo.srcObject = stream;

  socket = new WebSocket("ws://" + window.location.host + "/ws");

  socket.onopen = async () => {
    console.log("WebSocket connected");

    setupWebRTC(stream);
  };

  socket.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    if (msg.type === "answer") {
      console.log("Received Answer");
      pc.setRemoteDescription({ type: "answer", sdp: msg.data });
    }
  };

  const onTrack = (event) => {
    console.log("Track received:", event.streams[0]);
    remoteVideo.srcObject = event.streams[0];
  };

  async function setupWebRTC(stream) {
    pc = new RTCPeerConnection({
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
    });

    pc.ontrack = onTrack;
    stream.getTracks().forEach((track) => pc.addTrack(track, stream));

    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);

    await new Promise((resolve) => {
      if (pc.iceGatheringState === "complete") resolve();
      else
        pc.onicecandidate = (e) => {
          if (!e.candidate) resolve();
        };
    });

    const msg = {
      type: "offer",
      data: pc.localDescription.sdp,
    };
    socket.send(JSON.stringify(msg));
    console.log("Offer sent via WebSocket");
  }
}

document.getElementById("startButton").onclick = start;

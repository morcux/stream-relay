let pc = null;
let socket = null;

const onTrack = (event) => {
  const stream = event.streams[0];
  console.log("New track received:", stream.id);

  if (document.getElementById(stream.id)) return;

  const videosDiv = document.getElementById("videos");

  const wrapper = document.createElement("div");
  wrapper.className = "video-wrapper";

  const videoEl = document.createElement("video");
  videoEl.id = stream.id;
  videoEl.autoplay = true;
  videoEl.playsInline = true;
  videoEl.srcObject = stream;

  const label = document.createElement("div");
  label.className = "label";
  label.innerText = "Remote Peer";

  wrapper.appendChild(videoEl);
  wrapper.appendChild(label);
  videosDiv.appendChild(wrapper);
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

  socket.send(
    JSON.stringify({
      type: "offer",
      data: pc.localDescription.sdp,
    }),
  );
  console.log("Initial Offer sent");
}

async function start() {
  const localVideo = document.getElementById("localVideo");

  const stream = await navigator.mediaDevices.getUserMedia({
    video: true,
    audio: false,
  });
  localVideo.srcObject = stream;

  socket = new WebSocket("ws://" + window.location.host + "/ws");

  socket.onopen = () => {
    console.log("WebSocket connected");
    setupWebRTC(stream);
  };

  socket.onmessage = async (event) => {
    const msg = JSON.parse(event.data);

    if (msg.type === "answer") {
      console.log("Received Answer");
      await pc.setRemoteDescription({ type: "answer", sdp: msg.data });
    } else if (msg.type === "offer") {
      console.log("Received Renegotiation Offer");

      await pc.setRemoteDescription({ type: "offer", sdp: msg.data });
      const answer = await pc.createAnswer();
      await pc.setLocalDescription(answer);


      socket.send(
        JSON.stringify({
          type: "answer",
          data: pc.localDescription.sdp,
        }),
      );
    }
  };
}

document.getElementById("startButton").onclick = start;

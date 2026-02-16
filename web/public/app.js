let pc = null;
async function start() {
  const localVideo = document.getElementById("localVideo");
  const remoteVideo = document.getElementById("remoteVideo");

  const stream = await navigator.mediaDevices.getUserMedia({
    video: true,
    audio: false,
  });
  localVideo.srcObject = stream;

  pc = new RTCPeerConnection({
    iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
  });

  pc.ontrack = (event) => {
    console.log("remote stream:", event.streams[0]);
    remoteVideo.srcObject = event.streams[0];
  };

  stream.getTracks().forEach((track) => pc.addTrack(track, stream));

  // TODO: Trickle ICE
  const offer = await pc.createOffer();
  await pc.setLocalDescription(offer);

  await new Promise((resolve) => {
    if (pc.iceGatheringState === "complete") {
      resolve();
    } else {
      pc.onicecandidate = (e) => {
        if (e.candidate === null) {
          resolve();
        }
      };
    }
  });

  const response = await fetch("/sdp", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(pc.localDescription),
  });

  const answer = await response.json();
  await pc.setRemoteDescription(answer);
}

document.getElementById("startButton").onclick = start;

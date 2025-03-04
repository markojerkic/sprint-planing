document
  .getElementById("join-room-form")
  .addEventListener("submit", function (event) {
    event.preventDefault();
    const formData = new FormData(event.target);
    const roomId = formData.get("roomId");
    const url = `/room/${roomId}`;
    window.location = url;
  });

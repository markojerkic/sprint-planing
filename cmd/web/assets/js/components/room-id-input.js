class RoomIdInputElement extends HTMLElement {
    constructor() {
        super();
        const roomId = document.querySelector("[data-room-id]").dataset.roomId;
        this.innerHTML = `
      <input type="hidden" name="roomId"  value="${roomId}" />
    `;
    }
}

customElements.define("room-id-input", RoomIdInputElement);

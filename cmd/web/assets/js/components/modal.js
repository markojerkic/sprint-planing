class ModalElement extends HTMLElement {
	constructor() {
		super();
	}

	connectedCallback() {
		this.render();
	}

	render() {
		const buttonName = this.getAttribute("buttonName");
		const modalTitle = this.getAttribute("modalTitle") ?? buttonName;
		const randomId = `modal-${Math.random().toString(36).substring(7)}`;
		this.innerHTML = `
        <button type="button"
            class="btn btn-blue-700"
            popovertarget="${randomId}"
            popovertargetaction="show">
            ${buttonName}
        </button>

        <div id="${randomId}"
            popover="manual"
            class="popover-container">
            <div class="popover-content">
                <div class="popover-header">
                    <h2>${modalTitle}</h2>
                    <button class="close-btn"
                        popovertarget="${randomId}"
                        popovertargetaction="hide">
                        &times;
                    </button>
                </div>
                <div class="popover-body">
                    <slot></slot>
                </div>
            </div>
        </div>

        `;
	}
}

customElements.define("ui-modal", ModalElement);

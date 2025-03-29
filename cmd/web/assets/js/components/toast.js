class Toast extends HTMLElement {
	constructor() {
		super();
		this.attachShadow({ mode: "open" });
	}

	connectedCallback() {
		this.render();
		setTimeout(() => {
			this.remove();
		}, 5000);
	}

	render() {
		this.shadowRoot.innerHTML = `
            <style>
                .toast {
                    background-color: var(--color-primary-light);
                    background-color: var(--color-card-bg);
                    color: var(--color-text-light);
                    border-radius: 12px;
                    border-left: 5px solid var(--color-primary-light);
                    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
                    padding: 1rem;
                    margin-bottom: 0.5rem;
                }
            </style>
            <div class="toast">
                <slot></slot>
            </div>
        `;
	}
}

class ToastContainer extends HTMLElement {
	constructor() {
		super();
		this.attachShadow({ mode: "open" });
	}

	connectedCallback() {
		this.render();

		document.addEventListener(
			"toast",
			/** @param {ToastEvent} e */
			(e) => {
				const toast = document.createElement("ui-toast");
				console.log("toast event", e);
				toast.textContent = e.detail.message;
				this.appendChild(toast);
			},
		);
	}

	render() {
		this.shadowRoot.innerHTML = `
            <style>
                .toast-container {
                    position: fixed;
                    bottom: 0;
                    right: 0;
                    padding: 1rem;
                    z-index: 1000;
                }
            </style>
            <div class="toast-container">
                <slot></slot>
            </div>
        `;
	}
}

class ToastEvent extends Event {
	/**
	 * @typedef {Object} ToastEventData
	 * @property {ToastDetail} detail
	 */

	/**
	 * @typedef {Object} ToastDetail
	 * @property {string} message
	 */

	/**
	 * @param {ToastEventData} detail
	 */
	constructor(detail) {
		super("toast");
		this.message = detail.detail.message;
		this.detail = detail;
	}
}

customElements.define("ui-toast", Toast);
customElements.define("ui-toast-container", ToastContainer);

class Toast extends HTMLElement {
	constructor() {
		super();
		this.attachShadow({ mode: "open" });
	}

	connectedCallback() {
		this.render();

		// Set up event listener for close button
		const closeBtn = this.shadowRoot.querySelector(".close-btn");
		if (closeBtn) {
			closeBtn.addEventListener("click", () => this.startExitAnimation());
		}

		// Set a timeout to start the exit animation
		setTimeout(() => {
			this.startExitAnimation();
		}, 4000);
	}

	startExitAnimation() {
		const toast = this.shadowRoot.querySelector(".toast");
		if (toast) {
			toast.classList.add("exit");

			// Remove after the animation duration
			setTimeout(() => {
				this.remove();
			}, 500); // Match this with your CSS transition time
		}
	}

	render() {
		const level = this.getAttribute("level") ?? "info";
		const bgColor = level === "info" ? "--color-card-bg" : "--color-error";
		const borderColor =
			level === "info" ? "--color-primary-light" : "--color-error-dark";

		this.shadowRoot.innerHTML = `
            <style>
                .toast {
                    background-color: var(--color-primary-light);
                    background-color: var(${bgColor});
                    color: var(--color-text-light);
                    border-radius: 12px;
                    border-left: 5px solid var(${borderColor});
                    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
                    padding: 1rem;
                    margin-bottom: 0.5rem;
                    position: relative; /* Add this for positioning the close button */

                    /* Animation properties */
                    transform: translateX(0);
                    opacity: 1;
                    animation: slide-in 0.3s ease;
                    transition: transform 0.5s ease, opacity 0.5s ease;
                }

                @keyframes slide-in {
                    from {
                        transform: translateX(100%);
                        opacity: 0;
                    }
                    to {
                        transform: translateX(0);
                        opacity: 1;
                    }
                }
                .close-btn {
                    background: none;
                    border: none;
                    color: var(--color-text-light, #999);
                    font-size: 1.5rem;
                    font-weight: bold;
                    cursor: pointer;
                    padding: 0;
                    line-height: 1;
                    transition: color 0.2s;
                    position: absolute;
                    top: 0.5rem;
                    right: 0.5rem;
                }

                .close-btn:hover {
                    color: var(--color-primary);
                }

                .toast.exit {
                    transform: translateX(100%);
                    opacity: 0;
                }

                .toast-content {
                    padding-right: 1.5rem; /* Make space for the close button */
                }
            </style>
            <div class="toast">
                <button class="close-btn">&times;</button>
                <div class="toast-content">
                    <slot></slot>
                </div>
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
				toast.setAttribute("level", e.detail.level);
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
                    display: flex;
                    flex-direction: column;
                    align-items: flex-end;
                }

                ::slotted(ui-toast) {
                    width: 100%;
                    max-width: 300px;
                }
            </style>
            <div class="toast-container">
                <slot></slot>
            </div>
        `;
	}
}

/**
 * @property {string} message
 * @property {string} level
 * @property {ToastDetail} detail
 */
class ToastEvent extends Event {
	/**
	 * @typedef {Object} ToastEventData
	 * @property {string} message
	 * @property {string} level
	 */
	/**
	 * @param {ToastEventData} detail
	 */
	constructor(detail) {
		super("toast", { bubbles: true }); // Add bubbles: true to ensure event propagation
		this.message = detail.message;
		this.level = detail.level;
		this.detail = detail;
	}
}

customElements.define("ui-toast", Toast);
customElements.define("ui-toast-container", ToastContainer);

class ModalElement extends HTMLElement {
	constructor() {
		super();
		this.attachShadow({ mode: "open" });
	}
	connectedCallback() {
		this.render();
		this.setupEventListeners();
	}

	setupEventListeners() {
		/** @type {HTMLButtonElement} */
		const popover = this.shadowRoot.querySelector("[popover]");

		popover?.addEventListener(
			"beforetoggle",
			/** @param {ToggleEvent} event */
			(event) => {
				if (event.newState === "open") {
					this.showBackdrop();
					setTimeout(() => {
						const slottedElements = this.shadowRoot
							.querySelector("slot")
							.assignedNodes({ flatten: true });
						const focusableSelector =
							'button, input:not([type="hidden"]), select, textarea, [tabindex]:not([tabindex="-1"])';
						const firstFocusable =
							[...slottedElements]
								// @ts-ignore
								.find(
									/** @param {HTMLElement} node */
									(node) => node.querySelector?.(focusableSelector),
								)
								?.querySelector(focusableSelector) ||
							this.shadowRoot.querySelector("input");
						firstFocusable?.focus();
					}, 50);
				} else {
					this.hideBackdrop();
				}
			},
		);

		// Escape key to close the modal
		popover?.addEventListener("keydown", (event) => {
			if (event.key === "Escape") {
				popover?.hidePopover();
			}
		});

		// Add event listener for closing the modal
		document.addEventListener("closemodal", () => {
			popover?.hidePopover();
		});
	}

	showBackdrop() {
		/** @type {HTMLDivElement} */
		let backdrop = document.querySelector(".modal-backdrop");
		if (!backdrop) {
			backdrop = document.createElement("div");
			backdrop.className = "modal-backdrop";
			document.body.appendChild(backdrop);
		}
		backdrop.style.display = "block";
	}

	hideBackdrop() {
		/** @type {HTMLDivElement} */
		const backdrop = document.querySelector(".modal-backdrop");
		if (backdrop) {
			backdrop.style.display = "none";
		}
	}

	render() {
		const buttonName = this.getAttribute("buttonName");
		const modalTitle = this.getAttribute("modalTitle") ?? buttonName;
		const buttonColor =
			this.getAttribute("buttonColor") ?? "var(--color-primary)";
		this._randomId = `modal-${Math.random().toString(36).substring(7)}`;
		this.shadowRoot.innerHTML = `
        ${this.createStyles(buttonColor).outerHTML}
        <button type="button"
            class="btn"
            popovertarget="${this._randomId}"
            popovertargetaction="show">
            ${buttonName}
        </button>
        <div id="${this._randomId}"
            popover="manual"
            class="popover-container">
            <div class="popover-content">
                <div class="popover-header">
                    <h2>${modalTitle}</h2>
                    <button class="close-btn"
                        popovertarget="${this._randomId}"
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

	/** @param {string} buttonColor */
	createStyles(buttonColor) {
		const style = document.createElement("style");
		style.textContent = `
            /* Global styles for backdrop */
            :host {
                --backdrop-color: rgba(0, 0, 0, 0.5);
            }
              [popover] {
                margin: 0;
                padding: 0;
                width: 100%;
                max-width: 500px;
                max-height: 80vh;
                overflow: auto;
                border: none;
                border-radius: 12px;
                box-shadow: 0 5px 20px rgba(0, 0, 0, 0.3);
                animation: popoverFadeIn 0.2s ease-out;
                z-index: 1000;

                /* Center the popover */
                position: fixed;
                top: 50%;
                left: 50%;
                transform: translate(-50%, -50%);

                /* Ensure position is set before animation starts */
                will-change: transform, opacity;
                transform-origin: center center;
              }

            /* Button styles */
            .btn {
                display: inline-block;
                padding: 0.875rem 1.5rem;
                background-color: ${buttonColor};
                color: white;
                border: none;
                border-radius: 8px;
                font-size: 1rem;
                font-weight: 600;
                cursor: pointer;
                transition: all 0.3s ease;
                text-align: center;
                text-decoration: none;
                height: 100%;
            }

            .btn:hover {
                background-color: var(--color-primary-dark, #1a1a1a);
            }

            /* Popover container styles */
            .popover-container {
                background-color: transparent;
                border: none !important; /* Remove default border */
                padding: 0 !important; /* Remove default padding */
                margin: 0 !important; /* Reset margins */
                box-shadow: none !important; /* Remove default shadow */
                max-width: none !important; /* Override max-width */
            }

            .popover-content {
                background-color: var(--color-card-bg, white);
                border-radius: 12px;
                border-left: 5px solid var(--color-primary, ${buttonColor});
                overflow: hidden;
                box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
                position: relative;
                z-index: 1001;
                max-width: 50rem;
                margin: 0 auto;
                animation: fadeIn 0.2s ease-out;
            }

            @keyframes fadeIn {
                from { opacity: 0; transform: translateY(-20px); }
                to { opacity: 1; transform: translateY(0); }
            }

            .popover-header {
                padding: 1.5rem;
                border-bottom: 1px solid var(--color-border-color, #eee);
                display: flex;
                justify-content: space-between;
                align-items: center;
            }

            .popover-header h2 {
                margin: 0;
                color: var(--color-primary, ${buttonColor});
                font-size: 1.5rem;
            }

            .popover-body {
                padding: 1.5rem;
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
            }

            .close-btn:hover {
                color: var(--color-primary, ${buttonColor});
            }
        `;
		return style;
	}
}

class CloseModalEvent extends Event {
	constructor() {
		super("closemodal", { bubbles: true, composed: true });
	}
}

customElements.define("ui-modal", ModalElement);

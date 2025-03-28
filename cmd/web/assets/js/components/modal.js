class ModalElement extends HTMLElement {
	constructor() {
		super();
		this.attachShadow({ mode: "open" });
	}

	connectedCallback() {
		this.render();
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
            }
            .popover-container {
              background-color: transparent;
            }

            .popover-content {
              background-color: var(--color-card-bg);
              border-radius: 12px;
              border-left: 5px solid var(--color-primary);
              overflow: hidden;
            }

            .popover-header {
              padding: 1.5rem;
              border-bottom: 1px solid var(--color-border-color);
              display: flex;
              justify-content: space-between;
              align-items: center;
            }

            .popover-header h2 {
              margin: 0;
              color: var(--color-primary);
              font-size: 1.5rem;
            }

            .popover-body {
              padding: 1.5rem;
            }

            .close-btn {
              background: none;
              border: none;
              color: var(--color-text-light);
              font-size: 1.5rem;
              font-weight: bold;
              cursor: pointer;
              padding: 0;
              line-height: 1;
              transition: color 0.2s;
            }

            .close-btn:hover {
              color: var(--color-primary);
            }
        `;
		return style;
	}
}

customElements.define("ui-modal", ModalElement);

/**
 * FlashingDiv adds a border to the div and removes it after a specified time.
 * @class FlashingDiv
 */
export class FlashingDiv extends HTMLElement {
	constructor() {
		super();
		this.attachShadow({ mode: "open" });
	}

	connectedCallback() {
		this.render();
		if (this.hasAttribute("flash")) {
			this.flash();
		}
	}

	flash() {
		const color = this.getAttribute("color") || "var(--color-primary-light)";
		const div = this.#firstElement();
		div.style.display = "";
		div.scrollIntoView({
			behavior: "smooth",
			block: "start",
		});

		const previousBorder = div.style.border;
		div.style.border = `2px solid ${color}`;
		setTimeout(() => {
			div.style.border = previousBorder;
		}, 3_000);
	}

	render() {
		this.shadowRoot.innerHTML = `
                <slot></slot>
        `;
	}

	/** @returns {HTMLElement} */
	#firstElement() {
		const slottedElements = this.shadowRoot
			.querySelector("slot")
			.assignedNodes({ flatten: true });
		if (slottedElements[0] instanceof HTMLElement) {
			return slottedElements[0];
		}
		throw new Error("No slotted elements found");
	}
}

customElements.define("ui-flashing-div", FlashingDiv);

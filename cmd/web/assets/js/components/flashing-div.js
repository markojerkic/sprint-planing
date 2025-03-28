/**
 * FlashingDiv adds a border to the div and removes it after a specified time.
 * @class FlashingDiv
 */
class FlashingDiv extends HTMLElement {
	constructor() {
		super();
		this.attachShadow({ mode: "open" });
	}

	connectedCallback() {
		this.render();

		const color = this.getAttribute("color") || "var(--color-primary-light)";
		const div = this.firstElement();
		div.scrollIntoView({ behavior: "smooth" });
		const previousBorder = div.style.border;
		div.style.border = `2px solid ${color}`;
		setTimeout(() => {
			div.style.border = previousBorder;
		}, 3_000);
	}

	/** @returns {HTMLElement} */
	firstElement() {
		const slottedElements = this.shadowRoot
			.querySelector("slot")
			.assignedNodes({ flatten: true });

		return slottedElements[0];
	}

	render() {
		this.shadowRoot.innerHTML = `
                <slot></slot>
        `;
	}
}

customElements.define("ui-flashing-div", FlashingDiv);

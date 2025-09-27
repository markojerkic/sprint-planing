class LineClamp extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: "open" });
        this.expanded = false;
        this.hasOverflow = false;
    }

    connectedCallback() {
        this.render();
        // Use ResizeObserver to detect changes in content size
        this.resizeObserver = new ResizeObserver(() => this.checkOverflow());
        this.resizeObserver.observe(this);

        // Initial check after the element is added to the DOM
        requestAnimationFrame(() => {
            this.checkOverflow();
        });
    }

    disconnectedCallback() {
        if (this.resizeObserver) {
            this.resizeObserver.disconnect();
        }
    }

    render() {
        this.shadowRoot.innerHTML = `
      <style>
        :host {
          display: block;
        }

        :host(.line-clamp) .content {
          overflow: hidden;
          display: -webkit-box;
          -webkit-box-orient: vertical;
          -webkit-line-clamp: 3;
        }

        .button-container {
          margin-top: 10px;
          display: none;
        }

        :host(.has-overflow) .button-container {
          display: block;
        }

        button {
          background: none;
          border: none;
          color: var(--color-primary-light, oklch(0.681 0.152 313.1deg));
          cursor: pointer;
          padding: 0;
          font-size: 0.875rem;
          font-weight: 600;
          font-family: "JetBrains Mono", monospace;
          transition: color 0.2s;
          text-decoration: none;
        }

        button:hover {
          color: var(--color-primary-dark, oklch(0.482 0.166 310.5deg));
          text-decoration: underline;
        }

        button:focus {
          outline: none;
          text-decoration: underline;
        }
      </style>
      <div class="content">
        <slot></slot>
      </div>
      <div class="button-container">
        <button class="toggle-button">
          ${this.expanded ? "Display less..." : "Display more..."}
        </button>
      </div>
    `;

        // Set up event listeners
        const button = this.shadowRoot.querySelector(".toggle-button");
        button?.addEventListener("click", () => this.toggleExpansion());
    }

    checkOverflow() {
        const contentDiv = this.shadowRoot.querySelector(".content");
        if (!contentDiv) return;

        // Get the slotted content
        const slot = this.shadowRoot.querySelector("slot");
        if (!slot) return;

        const slottedNodes = slot.assignedNodes({ flatten: true });

        // Get the total text content height
        let totalHeight = 0;
        slottedNodes.forEach((node) => {
            if (node.nodeType === Node.ELEMENT_NODE) {
                totalHeight += node.offsetHeight;
            } else if (
                node.nodeType === Node.TEXT_NODE &&
                node.textContent.trim()
            ) {
                // For text nodes, we need a different approach
                const range = document.createRange();
                range.selectNodeContents(node);
                const rect = range.getBoundingClientRect();
                totalHeight += rect.height;
            }
        });

        // Apply the non-clamped state to measure full height
        this.classList.remove("line-clamp");
        const fullHeight = totalHeight;

        // Get computed line height for estimating number of lines
        const computedStyle = getComputedStyle(this);
        const lineHeight =
            parseInt(computedStyle.lineHeight) ||
            parseInt(computedStyle.fontSize) * 1.2; // Fallback estimate

        // Calculate approximate number of lines
        const lines = Math.ceil(fullHeight / lineHeight);

        this.hasOverflow = lines > 3;

        // Update classes based on overflow state
        if (this.hasOverflow) {
            this.classList.add("has-overflow");
            if (!this.expanded) {
                this.classList.add("line-clamp");
            }
        } else {
            this.classList.remove("has-overflow", "line-clamp");
        }

        // Update button text
        const button = this.shadowRoot.querySelector(".toggle-button");
        if (button) {
            button.textContent = this.expanded
                ? "Display less..."
                : "Display more...";
        }
    }

    toggleExpansion() {
        this.expanded = !this.expanded;

        if (this.expanded) {
            this.classList.remove("line-clamp");
        } else {
            this.classList.add("line-clamp");
        }

        // Update button text
        const button = this.shadowRoot.querySelector(".toggle-button");
        if (button) {
            button.textContent = this.expanded
                ? "Display less..."
                : "Display more...";
        }
    }
}

customElements.define("ui-line-clamp", LineClamp);

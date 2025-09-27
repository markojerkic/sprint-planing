/**
 * @fileoverview A web component that displays a progress bar at the top of the screen
 * during HTMX requests, particularly for boosted links.
 * @author Claude
 * @version 1.0.0
 */

/**
 * CSS variables that match the theme
 * @type {Object}
 */
const THEME = {
    PRIMARY: "oklch(0.591 0.175 311.5deg)",
    PRIMARY_DARK: "oklch(0.482 0.166 310.5deg)",
    PRIMARY_LIGHT: "oklch(0.681 0.152 313.1deg)",
    SUCCESS: "oklch(0.673 0.157 145.1deg)",
    WARNING: "oklch(0.77 0.173 63.6deg)",
    ERROR: "oklch(0.643 0.215 28deg)",
};

/**
 * Template for the HtmxProgressBar component
 * @returns {HTMLTemplateElement}
 */
function createTemplate() {
    const template = document.createElement("template");
    template.innerHTML = `
    <style>
      :host {
        display: block;
        position: relative;
      }

      .progress-container {
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 4px;
        background: transparent;
        z-index: 9999;
        pointer-events: none;
        transition: opacity 0.3s ease;
        opacity: 0;
      }

      .progress-container.active {
        opacity: 1;
      }

      .progress-bar {
        height: 100%;
        width: 0;
        background: linear-gradient(to right, ${THEME.PRIMARY}, ${THEME.PRIMARY_LIGHT});
        transition: width 0.1s ease;
        position: relative;
        overflow: hidden;
      }

      .progress-bar.error {
        background: linear-gradient(to right, ${THEME.ERROR}, ${THEME.ERROR});
      }

      .progress-bar::after {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        bottom: 0;
        right: 0;
        background: linear-gradient(
          to right,
          rgba(255, 255, 255, 0),
          rgba(255, 255, 255, 0.3),
          rgba(255, 255, 255, 0)
        );
        transform: translateX(-100%);
        animation: shimmer 2s infinite;
      }

      @keyframes shimmer {
        100% {
          transform: translateX(100%);
        }
      }
    </style>

    <div class="progress-container">
      <div class="progress-bar"></div>
    </div>
  `;
    return template;
}

/**
 * HtmxProgressBar web component class
 * @class
 * @extends {HTMLElement}
 */
class HtmxProgressBar extends HTMLElement {
    /**
     * Creates an instance of HtmxProgressBar
     * @constructor
     */
    constructor() {
        super();

        /**
         * Shadow DOM for the component
         * @type {ShadowRoot}
         * @private
         */
        this.attachShadow({ mode: "open" });

        // Clone template content into shadow DOM
        this.shadowRoot.appendChild(createTemplate().content.cloneNode(true));

        /**
         * Reference to the progress container element
         * @type {HTMLElement}
         * @private
         */
        this.progressContainer = this.shadowRoot.querySelector(
            ".progress-container",
        );

        /**
         * Reference to the progress bar element
         * @type {HTMLElement}
         * @private
         */
        this.progressBar = this.shadowRoot.querySelector(".progress-bar");

        /**
         * Count of active requests
         * @type {number}
         * @private
         */
        this.activeRequests = 0;

        /**
         * Current progress value (0-100)
         * @type {number}
         * @private
         */
        this.progressValue = 0;

        /**
         * Interval for progress animation
         * @type {number|null}
         * @private
         */
        this.progressInterval = null;

        /**
         * Error state flag
         * @type {boolean}
         * @private
         */
        this.hasError = false;

        // Bind event handlers
        this.handleBeforeRequest = this.handleBeforeRequest.bind(this);
        this.handleAfterRequest = this.handleAfterRequest.bind(this);
        this.handleRequestError = this.handleRequestError.bind(this);

        // Initialize attributes and styles
        this.initializeAttributes();
    }

    /**
     * Called when the element is added to the DOM
     * Sets up event listeners for HTMX events
     */
    connectedCallback() {
        // Register HTMX event listeners
        document.addEventListener(
            "htmx:beforeRequest",
            this.handleBeforeRequest,
        );
        document.addEventListener("htmx:afterRequest", this.handleAfterRequest);
        document.addEventListener(
            "htmx:responseError",
            this.handleRequestError,
        );
    }

    /**
     * Called when the element is removed from the DOM
     * Cleans up event listeners
     */
    disconnectedCallback() {
        document.removeEventListener(
            "htmx:beforeRequest",
            this.handleBeforeRequest,
        );
        document.removeEventListener(
            "htmx:afterRequest",
            this.handleAfterRequest,
        );
        document.removeEventListener(
            "htmx:responseError",
            this.handleRequestError,
        );
        this.stopProgressAnimation();
    }

    /**
     * Specifies which attributes to observe for changes
     * @returns {string[]} Array of attribute names to observe
     * @static
     */
    static get observedAttributes() {
        return ["color", "height", "boosted-only", "error-color"];
    }

    /**
     * Called when an observed attribute changes
     * @param {string} name - Name of the attribute
     * @param {string} oldValue - Previous value
     * @param {string} newValue - New value
     */
    attributeChangedCallback(name, oldValue, newValue) {
        if (oldValue === newValue) return;

        switch (name) {
            case "color":
                this.updateProgressBarColor(newValue);
                break;
            case "height":
                this.progressContainer.style.height = newValue;
                break;
            case "error-color":
                // Store for later use when errors occur
                this.errorColor = newValue;
                break;
        }
    }

    /**
     * Initializes component attributes from HTML attributes
     * @private
     */
    initializeAttributes() {
        // Set default or custom colors
        this.updateProgressBarColor(
            this.getAttribute("color") || THEME.PRIMARY,
        );

        // Set error color
        this.errorColor = this.getAttribute("error-color") || THEME.ERROR;

        // Set height
        const height = this.getAttribute("height");
        if (height) {
            this.progressContainer.style.height = height;
        }
    }

    /**
     * Updates the progress bar's color
     * @param {string} color - CSS color value
     * @private
     */
    updateProgressBarColor(color) {
        if (!color) return;

        // If color is a theme variable, use that
        const themeColor = color.startsWith("--") ? `var(${color})` : color;

        this.progressBar.style.background = `linear-gradient(to right, ${themeColor}, ${themeColor})`;
    }

    /**
     * Handles the start of an HTMX request
     * @param {CustomEvent} event - The htmx:beforeRequest event
     * @private
     */
    handleBeforeRequest(event) {
        // Check if we should only track boosted requests
        const trackOnlyBoosted = this.hasAttribute("boosted-only");
        const isBoosted =
            event.detail.elt.getAttribute("hx-boost") === "true" ||
            (event.detail.requestConfig && event.detail.requestConfig.boosted);

        if (trackOnlyBoosted && !isBoosted) {
            return;
        }

        this.activeRequests++;

        if (this.activeRequests === 1) {
            // Reset error state
            this.hasError = false;
            this.progressBar.classList.remove("error");

            // Update color back to normal
            this.updateProgressBarColor(
                this.getAttribute("color") || THEME.PRIMARY,
            );

            // First active request, show progress bar
            this.progressContainer.classList.add("active");
            this.startProgressAnimation();
        }
    }

    /**
     * Handles the completion of an HTMX request
     * @param {CustomEvent} event - The htmx:afterRequest event
     * @private
     */
    handleAfterRequest(event) {
        // Check if we should only track boosted requests
        const trackOnlyBoosted = this.hasAttribute("boosted-only");
        const isBoosted =
            event.detail.elt.getAttribute("hx-boost") === "true" ||
            (event.detail.requestConfig && event.detail.requestConfig.boosted);

        if (trackOnlyBoosted && !isBoosted) {
            return;
        }

        if (this.activeRequests > 0) {
            this.activeRequests--;
        }

        if (this.activeRequests === 0) {
            // No more active requests, complete the progress bar
            this.completeProgress();
        }
    }

    /**
     * Handles HTMX request errors
     * @param {CustomEvent} event - The htmx:responseError event
     * @private
     */
    handleRequestError(event) {
        // Mark as error
        this.hasError = true;
        this.progressBar.classList.add("error");

        // Update to error color
        this.updateProgressBarColor(this.errorColor);

        // Complete the request normally
        this.handleAfterRequest(event);
    }

    /**
     * Starts the progress animation
     * @private
     */
    startProgressAnimation() {
        // Reset progress
        this.progressValue = 0;
        this.updateProgressBar();

        // Stop any existing interval
        this.stopProgressAnimation();

        // Simulate progress until request completes
        this.progressInterval = setInterval(() => {
            // Slow down progress as it gets closer to 90%
            if (this.progressValue < 90) {
                const increment = (90 - this.progressValue) / 10;
                this.progressValue += Math.max(0.5, increment);
                this.updateProgressBar();
            }
        }, 100);
    }

    /**
     * Stops the progress animation
     * @private
     */
    stopProgressAnimation() {
        if (this.progressInterval) {
            clearInterval(this.progressInterval);
            this.progressInterval = null;
        }
    }

    /**
     * Completes the progress bar animation
     * @private
     */
    completeProgress() {
        this.stopProgressAnimation();
        this.progressValue = 100;
        this.updateProgressBar();

        // Hide after completion animation finishes
        setTimeout(() => {
            this.progressContainer.classList.remove("active");

            // Reset after hiding
            setTimeout(() => {
                this.progressValue = 0;
                this.updateProgressBar();

                // Reset error state
                if (this.hasError) {
                    this.hasError = false;
                    this.progressBar.classList.remove("error");
                    this.updateProgressBarColor(
                        this.getAttribute("color") || THEME.PRIMARY,
                    );
                }
            }, 300);
        }, 200);
    }

    /**
     * Updates the progress bar width
     * @private
     */
    updateProgressBar() {
        this.progressBar.style.width = `${this.progressValue}%`;
    }
}

// Register the custom element
if (!customElements.get("ui-progress-bar")) {
    customElements.define("ui-progress-bar", HtmxProgressBar);
}

// Export the class for module usage
export default HtmxProgressBar;

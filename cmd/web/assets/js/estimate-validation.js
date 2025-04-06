/** @param {HTMLFormElement} form */
function validateEstimation(form) {
	/** @type {HTMLInputElement} */
	const weekInput = form.querySelector('input[name="weekEstimate"]');
	/** @type {HTMLInputElement} */
	const dayInput = form.querySelector('input[name="dayEstimate"]');
	/** @type {HTMLInputElement} */
	const hourInput = form.querySelector('input[name="hourEstimate"]');

	if (!weekInput || !dayInput || !hourInput) {
		console.error("Missing input fields");
		return false; // Prevent form submission
	}

	// Add event listeners to clear validation when user inputs new values
	if (!weekInput.dataset.validationInitialized) {
		weekInput.addEventListener("input", clearValidationError);
		dayInput.addEventListener("input", clearValidationError);
		hourInput.addEventListener("input", clearValidationError);

		// Mark as initialized to avoid adding listeners multiple times
		weekInput.dataset.validationInitialized = "true";
		dayInput.dataset.validationInitialized = "true";
		hourInput.dataset.validationInitialized = "true";
	}

	const week = parseInt(weekInput.value) || 0;
	const day = parseInt(dayInput.value) || 0;
	const hour = parseInt(hourInput.value) || 0;

	// Reset previous validation errors
	weekInput.setCustomValidity("");

	// Check if all values are empty or zero
	if (week <= 0 && day <= 0 && hour <= 0) {
		weekInput.setCustomValidity(
			"Please enter at least one non-zero estimate value",
		);
		form.reportValidity();
		return false; // Prevent form submission
	}

	return true; // Allow form submission
}

/**
 * Clears validation errors when user inputs new values
 * @param {Event} event
 */
function clearValidationError(event) {
	/** @type {HTMLInputElement} input */
	const input = event.target;
	const form = input.closest("form");

	if (form) {
		/** @type {HTMLInputElement} */
		const weekInput = form.querySelector('input[name="weekEstimate"]');
		if (weekInput) {
			weekInput.setCustomValidity("");
		}
	}
}

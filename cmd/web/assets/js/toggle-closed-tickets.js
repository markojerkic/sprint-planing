function toggleClosedTickets() {
    localStorage.setItem(
        "hideClosedTickets",
        !document.body.classList.contains("hide-closed-tickets")
            ? "true"
            : "false",
    );
    setCurrentState();
}

function setCurrentState() {
    const isHidden = localStorage.getItem("hideClosedTickets") === "true";
    const toggleButton = document.getElementById("toggle-hidden-tickets");
    if (!toggleButton) {
        console.error("Toggle button not found");
        return;
    }
    if (isHidden) {
        document.body.classList.add("hide-closed-tickets");
        toggleButton.textContent = "Show Closed Tickets";
        toggleButton.classList.add("btn-sm-success");
        toggleButton.classList.remove("btn-sm-warning");
    } else {
        document.body.classList.remove("hide-closed-tickets");
        toggleButton.textContent = "Hide Closed Tickets";
        toggleButton.classList.add("btn-sm-warning");
        toggleButton.classList.remove("btn-sm-success");
    }
}

document.addEventListener("DOMContentLoaded", () => {
    setCurrentState();
});
document.addEventListener("htmx:afterSwap", (e) => {
    const isBoosted = e.detail.requestConfig?.boosted;
    if (isBoosted) {
        setCurrentState();
    }
});

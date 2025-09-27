let wsHasConnectedBefore = false;
let isRefetching = false;

htmx.on("htmx:wsOpen", async function () {
    if (!wsHasConnectedBefore) {
        wsHasConnectedBefore = true;
        return;
    }
    if (isRefetching) {
        return;
    }

    while (hasFocusedInput()) {
        console.log("Waiting for input to be unfocused...");
        await new Promise((resolve) => setTimeout(resolve, 3_000));
    }

    console.log("Reconnecting...");
    htmx.ajax("get", window.location.pathname, {
        target: "#ticket-list",
        select: "#ticket-list",
    }).then(function () {
        isRefetching = false;
    });
});

function hasFocusedInput() {
    return document.activeElement.tagName === "INPUT";
}

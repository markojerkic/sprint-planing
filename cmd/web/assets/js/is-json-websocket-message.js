/**
 * @typedef {Object} Message
 * @property {string} messageType
 * @property {any} data
 *
 * @param {string} rawJson
 * @param {string} type
 * @returns {Message|null}
 */
function isJsonWebSocketMessage(rawJson, type) {
    if (typeof rawJson !== "string") {
        return null;
    }

    if (!rawJson.startsWith("{") && !rawJson.startsWith("[")) {
        return null;
    }

    try {
        const json = JSON.parse(rawJson);
        if (json.messageType === type) {
            return json;
        }
    } catch (e) {
        return null;
    }
}

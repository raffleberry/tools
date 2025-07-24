chrome.commands.onCommand.addListener(function (command) {
  if (command === "autofill-text") {
    chrome.storage.local.get("autofillText", ({ autofillText }) => {
      if (!autofillText) {
        console.warn("No autofill text found.");
        return;
      }

      chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
        const tab = tabs[0];

        chrome.tabs.executeScript(tab.id, {
          code: `
            (function() {
              const el = document.activeElement;
              if (el && (el.tagName === "INPUT" || el.tagName === "TEXTAREA")) {
                el.value = ${JSON.stringify(autofillText)};
                el.dispatchEvent(new Event("input", { bubbles: true }));
              } else {
                alert("Focus a text input field first.");
              }
            })();
          `
        });
      });
    });
  }
});

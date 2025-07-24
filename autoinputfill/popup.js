const input = document.getElementById("autofillText");
const button = document.getElementById("fillBtn");

chrome.storage.local.get("autofillText", (data) => {
  if (data.autofillText) {
    input.value = data.autofillText;
  }
});

input.addEventListener("input", () => {
  chrome.storage.local.set({ autofillText: input.value });
});

button.addEventListener("click", () => {
  const text = input.value;

  chrome.storage.local.set({ autofillText: text }); // optional redundancy

  chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
    const tab = tabs[0];

    chrome.tabs.executeScript(tab.id, {
      code: `
        (function() {
          const el = document.activeElement;
          console.log(el);
          if (el && (el.tagName === "INPUT" || el.tagName === "TEXTAREA")) {
            el.value = ${JSON.stringify(text)};
            el.dispatchEvent(new Event("input", { bubbles: true }));
          } else {
            alert("Focus a text input field first.");
          }
        })();
      `
    });
  });
});

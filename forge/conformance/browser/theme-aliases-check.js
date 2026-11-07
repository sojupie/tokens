const button = document.querySelector("#run"),
    result = document.querySelector("#result");
  button.onclick = async () => {
    button.disabled = true;
    try {
      const [current, previous] = await Promise.all(
        ["/theme.css", "baseline-theme.tmp"].map((url) =>
          fetch(url).then((r) => {
            if (!r.ok) throw Error(url + ": " + r.status);
            return r.text();
          }),
        ),
      );
      const frame = document.querySelector("iframe");
      frame.srcdoc = "<html><head></head><body></body></html>";
      await new Promise((resolve) => (frame.onload = resolve));
      const d = frame.contentDocument,
        w = frame.contentWindow,
        style = d.createElement("style");
      d.head.append(style);
      const names = [
        ...new Set(
          [...current.matchAll(/(--[\w-]+):/g)].map((match) => match[1]),
        ),
      ];
      const differences = [];
      let checks = 0;
      for (const theme of ["light", "dark"])
        for (const contrast of ["default", "medium", "high"]) {
          d.documentElement.dataset.theme = theme;
          d.documentElement.dataset.contrast = contrast;
          style.textContent = current;
          const snapshot = names.map((name) =>
            w
              .getComputedStyle(d.documentElement)
              .getPropertyValue(name)
              .replace(/\s/g, ""),
          );
          style.textContent = previous;
          for (let i = 0; i < names.length; i++) {
            const expected = w
              .getComputedStyle(d.documentElement)
              .getPropertyValue(names[i])
              .replace(/\s/g, "");
            if (!expected) continue;
            checks++;
            if (snapshot[i] !== expected && differences.length < 100)
              differences.push({
                theme,
                contrast,
                name: names[i],
                before: expected,
                after: snapshot[i],
              });
          }
        }
      result.textContent = JSON.stringify({ checks, differences }, null, 2);
    } catch (error) {
      result.textContent = error.stack;
    } finally {
      button.disabled = false;
    }
  };

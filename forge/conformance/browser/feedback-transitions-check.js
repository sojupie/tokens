import { checkFeedbackTransitions } from "./check-feedback-transitions.mjs";
      const origin = "http://127.0.0.1:8081/md3/",
        version = Date.now();
      await Promise.all(
        [
          "theme.css",
          "static/app.css",
          "static/token-table.css",
          "static/buttons.css",
          "static/components.css",
        ].map(
          (path) =>
            new Promise((resolve, reject) => {
              const link = document.createElement("link");
              link.rel = "stylesheet";
              link.href = origin + path + "?check=" + version;
              link.onload = resolve;
              link.onerror = reject;
              document.head.append(link);
            }),
        ),
      );
      for (const name of ["motion", "buttons"]) await new Promise((resolve, reject) => {
        const script = document.createElement("script");
        script.src = origin + `static/${name}.js?check=` + version;
        script.onload = resolve;
        script.onerror = reject;
        document.head.append(script);
      });
      const run = document.getElementById("run"),
        output = document.getElementById("result");
      run.disabled = false;
      output.textContent = "Ready";
      run.addEventListener("click", async () => {
        run.disabled = true;
        output.textContent = "Running";
        const results = [];
        for (const theme of ["light", "dark"])
          for (const contrast of ["default", "medium", "high"]) {
            document.documentElement.dataset.theme = theme;
            document.documentElement.dataset.contrast = contrast;
            results.push({
              theme,
              contrast,
              checks: await checkFeedbackTransitions([
                ...document.querySelectorAll(
                  "#samples .md-interactive-card,#samples .md-icon-button,#samples .md-filter-chip,#samples .md-menu-item,#samples .md-button",
                ),
              ]),
            });
          }
        output.textContent = JSON.stringify(results, null, 2);
        run.disabled = false;
      });

(function() {
    let gothicframework_reloadSrc = window.gothicframework_reloadSrc || new EventSource("/_gothicframework/reload/events");
    gothicframework_reloadSrc.onmessage = (event) => {
      if (event && event.data === "reload") {
        window.location.reload();
      }
    };
    window.gothicframework_reloadSrc = gothicframework_reloadSrc;
    window.onbeforeunload = () => window.gothicframework_reloadSrc.close();
  })();
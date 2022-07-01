import React from "react";
import { createRoot } from "react-dom/client";

const apps = {
  InterestForm: () => import("./InterestForm.tsx"),
};

const renderAppInElement = (el) => {
  if (apps[el.id]) {
    apps[el.id]().then((module) => {
      const App = module.default;
      const root = createRoot(el);
      root.render(<App {...el.dataset} />);
    });
  }
};

document.querySelectorAll(".__react-root").forEach(renderAppInElement);

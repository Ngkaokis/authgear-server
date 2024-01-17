const cssVarsToAttrs = {
  "--alignment-logo": "alignment-logo",
  "--alignment-content": "alignment-content",
};

export function injectCSSAttrs(el: HTMLElement) {
  for (const [cssVar, attr] of Object.entries(cssVarsToAttrs)) {
    const value = getComputedStyle(el).getPropertyValue(cssVar);
    el.setAttribute(attr, value);
  }
}

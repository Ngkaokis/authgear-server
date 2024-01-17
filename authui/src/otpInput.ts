import { Controller } from "@hotwired/stimulus";

export class OtpInputController extends Controller {
  static targets = ["input", "submit", "digitsContainer"];

  declare readonly inputTarget: HTMLInputElement;
  declare readonly submitTarget: HTMLButtonElement;
  declare readonly digitsContainerTarget: HTMLElement;

  spans: HTMLElement[] = [];

  get maxLength(): number {
    if (
      this.inputTarget.maxLength != null &&
      this.inputTarget.maxLength !== 0
    ) {
      return this.inputTarget.maxLength;
    }
    return 6;
  }

  get value(): string {
    return this.inputTarget.value;
  }

  connect(): void {
    this.inputTarget.addEventListener("input", this.handleInput);
    this.inputTarget.addEventListener("paste", this.handlePaste);
    this.inputTarget.addEventListener("focus", this.handleFocus);
    this.inputTarget.addEventListener("blur", this.handleBlur);
    // element.selectionchange is NOT the same as document.selectionchange
    // element.selectionchange is an experimental technology.
    window.document.addEventListener(
      "selectionchange",
      this.handleSelectionChange
    );
    this.render();
  }

  disconnect(): void {
    this.inputTarget.removeEventListener("input", this.handleInput);
    this.inputTarget.removeEventListener("paste", this.handlePaste);
    this.inputTarget.removeEventListener("focus", this.handleFocus);
    this.inputTarget.removeEventListener("blur", this.handleBlur);
    window.document.removeEventListener(
      "selectionchange",
      this.handleSelectionChange
    );
  }

  _setValue = (value: string): void => {
    this.inputTarget.value = value
      .replace(/[^0-9]/g, "")
      .slice(0, this.maxLength);

    const reachedMaxDigits = this.value.length === this.maxLength;
    if (reachedMaxDigits) {
      this.submitTarget.click();
    }

    this.render();
  };

  handleInput = (event: Event): void => {
    const input = event.target as HTMLInputElement;
    this._setValue(input.value);
  };

  handlePaste = (event: ClipboardEvent): void => {
    event.preventDefault();
    const text = event.clipboardData?.getData("text/plain");
    if (text) {
      this._setValue(text);
    }
  };

  handleFocus = (event: FocusEvent): void => {
    const input = event.target as HTMLInputElement;
    input.setSelectionRange(input.value.length, input.value.length);
    this.render();
  };

  handleBlur = (): void => {
    this.render();
  };

  handleSelectionChange = (_event: Event): void => {
    if (this.inputTarget === document.activeElement) {
      this.inputTarget.setSelectionRange(
        this.inputTarget.value.length,
        this.inputTarget.value.length
      );
    }
  };

  isSpanSelected = (index: number): boolean => {
    const isFocused = this.inputTarget === document.activeElement;
    const isNextBox = this.value.length === index;
    return isFocused && isNextBox;
  };

  render = (): void => {
    const digitsContainer = this.digitsContainerTarget;
    if (this.spans.length !== this.maxLength) {
      digitsContainer.innerHTML = "";
    }

    for (let i = 0; i < this.maxLength; i++) {
      let textContent = this.value.slice(i, i + 1) || "";
      let className = this.isSpanSelected(i)
        ? "otp-input__digit otp-input__digit--focus"
        : "otp-input__digit";

      if (textContent !== "") {
        textContent = " ";
        className += " otp-input__digit--masked";
      }

      this.inputTarget.style.letterSpacing = `calc(${this.inputTarget.offsetWidth}px / ${this.maxLength})`;

      let span = this.spans[i];
      if (!span) {
        span = document.createElement("span");
        digitsContainer.appendChild(span);
        this.spans[i] = span;
      }

      span.textContent = textContent;
      span.className = className;
    }
  };
}

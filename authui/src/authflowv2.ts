import { start } from "@hotwired/turbo";
import { Application } from "@hotwired/stimulus";
import axios from "axios";
import { init as SentryInit } from "@sentry/browser";
import { BrowserTracing } from "@sentry/tracing";
import {
  RestoreFormController,
  RetainFormFormController,
  RetainFormInputController,
} from "./form";
import { XHRSubmitFormController } from "./authflowv2/form";
import { LoadingController } from "./authflowv2/loading";
import { PreventDoubleTapController } from "./preventDoubleTap";
import { LockoutController } from "./lockout";
import { FormatDateRelativeController } from "./date";
import { injectCSSAttrs } from "./cssattrs";
import { ResendButtonController } from "./resendButton";
import { OtpInputController } from "./authflowv2/otpInput";
import { PasswordVisibilityToggleController } from "./passwordVisibility";
import { PasswordPolicyController } from "./authflowv2/password-policy";
import { PasswordStrengthMeterController } from "./authflowv2/password-strength-meter";
import { PhoneInputController } from "./authflowv2/phoneInput";
import { CustomSelectController } from "./authflowv2/customSelect";
import { CountdownController } from "./countdown";
import { TextFieldController } from "./authflowv2/text-field";
import { DialogController } from "./authflowv2/dialog";
import { CopyButtonController } from "./copy";
import { AuthflowWebsocketController } from "./authflow_websocket";
import { AuthflowPollingController } from "./authflow_polling";
import {
  AuthflowPasskeyRequestController,
  AuthflowPasskeyCreationController,
} from "./passkey";
import { NewPasswordFieldController } from "./authflowv2/new-password-field";

axios.defaults.withCredentials = true;

const sentryDSN = document
  .querySelector("meta[name=x-sentry-dsn]")
  ?.getAttribute("content");
if (sentryDSN != null && sentryDSN !== "") {
  SentryInit({
    dsn: sentryDSN,
    integrations: [new BrowserTracing()],
    // Do not enable performance monitoring.
    // tracesSampleRate: 0,
  });
}
start();

const Stimulus = Application.start();

Stimulus.register("xhr-submit-form", XHRSubmitFormController);
Stimulus.register("restore-form", RestoreFormController);
Stimulus.register("retain-form-form", RetainFormFormController);
Stimulus.register("retain-form-input", RetainFormInputController);

Stimulus.register("prevent-double-tap", PreventDoubleTapController);

Stimulus.register("lockout", LockoutController);

Stimulus.register("format-date-relative", FormatDateRelativeController);
Stimulus.register("format-date-relative", FormatDateRelativeController);
Stimulus.register(
  "password-visibility-toggle",
  PasswordVisibilityToggleController
);

Stimulus.register("otp-input", OtpInputController);
Stimulus.register("resend-button", ResendButtonController);
Stimulus.register("password-policy", PasswordPolicyController);
Stimulus.register("password-strength-meter", PasswordStrengthMeterController);
Stimulus.register("custom-select", CustomSelectController);
Stimulus.register("phone-input", PhoneInputController);
Stimulus.register("countdown", CountdownController);
Stimulus.register("copy-button", CopyButtonController);

Stimulus.register("text-field", TextFieldController);
Stimulus.register("dialog", DialogController);
Stimulus.register("loading", LoadingController);
Stimulus.register("new-password-field", NewPasswordFieldController);

Stimulus.register("authflow-websocket", AuthflowWebsocketController);
Stimulus.register("authflow-polling", AuthflowPollingController);
Stimulus.register("authflow-passkey-request", AuthflowPasskeyRequestController);
Stimulus.register(
  "authflow-passkey-creation",
  AuthflowPasskeyCreationController
);

injectCSSAttrs(document.documentElement);

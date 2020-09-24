import React, {
  useMemo,
  useCallback,
  useContext,
  useState,
  useEffect,
} from "react";
import { useNavigate } from "react-router-dom";
import cn from "classnames";
import {
  DefaultButton,
  Dialog,
  DialogFooter,
  Icon,
  List,
  PrimaryButton,
  Text,
} from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";

import { useDeleteAuthenticatorMutation } from "./mutations/deleteAuthenticatorMutation";
import ListCellLayout from "../../ListCellLayout";
import ShowError from "../../ShowError";
import ButtonWithLoading from "../../ButtonWithLoading";
import { destructiveTheme } from "../../theme";
import { formatDatetime } from "../../util/formatDatetime";
import { parseError } from "../../util/error";
import {
  defaultFormatErrorMessageList,
  Violation,
} from "../../util/validation";

import styles from "./UserDetailsAccountSecurity.module.scss";

// authenticator type recognized by portal
type PrimaryAuthenticatorType = "PASSWORD" | "OOB_OTP";
type SecondaryAuthenticatorType = "PASSWORD" | "TOTP" | "OOB_OTP";
type AuthenticatorType = PrimaryAuthenticatorType | SecondaryAuthenticatorType;

type AuthenticatorKind = "PRIMARY" | "SECONDARY";

type OOBOTPVerificationMethod = "email" | "phone" | "unknown";

interface AuthenticatorClaims extends Record<string, unknown> {
  email?: string;
  phone_number?: string;
}

interface Authenticator {
  id: string;
  type: AuthenticatorType;
  kind: AuthenticatorKind;
  isDefault: boolean;
  claims: AuthenticatorClaims;
  createdAt: string;
  updatedAt: string;
}

interface UserDetailsAccountSecurityProps {
  authenticators: Authenticator[];
}

interface PasswordAuthenticatorData {
  id: string;
  kind: AuthenticatorKind;
  lastUpdated: string;
}

interface TOTPAuthenticatorData {
  id: string;
  kind: AuthenticatorKind;
  label: string;
  addedOn: string;
}

interface OOBOTPAuthenticatorData {
  id: string;
  iconName?: string;
  kind: AuthenticatorKind;
  label: string;
  addedOn: string;
  isDefault: boolean;
}

interface PasswordAuthenticatorCellProps extends PasswordAuthenticatorData {
  showConfirmationDialog: (
    authenticatorID: string,
    authenticatorName: string
  ) => void;
}

interface TOTPAuthenticatorCellProps extends TOTPAuthenticatorData {
  showConfirmationDialog: (
    authenticatorID: string,
    authenticatorName: string
  ) => void;
}

interface OOBOTPAuthenticatorCellProps extends OOBOTPAuthenticatorData {
  showConfirmationDialog: (
    authenticatorID: string,
    authenticatorName: string
  ) => void;
}

interface RemoveConfirmationDialogData {
  authenticatorID: string;
  authenticatorName: string;
}

interface RemoveConfirmationDialogProps
  extends Partial<RemoveConfirmationDialogData> {
  visible: boolean;
  deleteAuthenticator: (authenticatorID: string) => void;
  deletingAuthenticator: boolean;
  onDismiss: () => void;
}

interface ErrorDialogData {
  errorMessage: string;
}

interface ErrorDialogProps extends Partial<ErrorDialogData> {
  visible: boolean;
  onDismiss: () => void;
}

const LABEL_PLACEHOLDER = "---";

const primaryAuthenticatorTypeLocaleKeyMap: {
  [key in AuthenticatorType]?: string;
} = {
  PASSWORD: "UserDetails.account-security.primary.password",
  OOB_OTP: "UserDetails.account-security.primary.oob-otp",
};

const secondaryAuthenticatorTypeLocaleKeyMap: {
  [key in AuthenticatorType]?: string;
} = {
  PASSWORD: "UserDetails.account-security.secondary.password",
  TOTP: "UserDetails.account-security.secondary.totp",
  OOB_OTP: "UserDetails.account-security.secondary.oob-otp",
};

function getLocaleKeyWithAuthenticatorType(
  type: AuthenticatorType,
  kind: AuthenticatorKind
): string | undefined {
  switch (kind) {
    case "PRIMARY":
      return primaryAuthenticatorTypeLocaleKeyMap[type];
    case "SECONDARY":
      return secondaryAuthenticatorTypeLocaleKeyMap[type];
    default:
      return undefined;
  }
}

function constructPasswordAuthenticatorData(
  authenticator: Authenticator,
  locale: string
): PasswordAuthenticatorData {
  const lastUpdated = formatDatetime(locale, authenticator.updatedAt) ?? "";

  return {
    id: authenticator.id,
    kind: authenticator.kind,
    lastUpdated,
  };
}

function getTotpDisplayName(
  totpAuthenticatorClaims: Record<string, unknown>
): string {
  for (const [key, claim] of Object.entries(totpAuthenticatorClaims)) {
    if (
      key === "https://authgear.com/claims/totp/display_name" &&
      typeof claim === "string"
    ) {
      return claim;
    }
  }
  return LABEL_PLACEHOLDER;
}

function constructTotpAuthenticatorData(
  authenticator: Authenticator,
  locale: string
): TOTPAuthenticatorData {
  const addedOn = formatDatetime(locale, authenticator.createdAt) ?? "";
  const label = getTotpDisplayName(authenticator.claims);

  return {
    id: authenticator.id,
    kind: authenticator.kind,
    addedOn,
    label,
  };
}

function getOobOtpVerificationMethod(
  authenticator: Authenticator
): OOBOTPVerificationMethod {
  if (authenticator.claims.email != null) {
    return "email";
  }
  if (authenticator.claims.phone_number != null) {
    return "phone";
  }
  return "unknown";
}

const oobOtpVerificationMethodIconName: Partial<Record<
  OOBOTPVerificationMethod,
  string
>> = {
  email: "Mail",
  phone: "CellPhone",
};

function getOobOtpAuthenticatorLabel(
  authenticator: Authenticator,
  verificationMethod: OOBOTPVerificationMethod
) {
  switch (verificationMethod) {
    case "email":
      return authenticator.claims.email ?? "";
    case "phone":
      return authenticator.claims.phone_number ?? "";
    default:
      return "";
  }
}

function constructOobOtpAuthenticatorData(
  authenticator: Authenticator,
  locale: string
): OOBOTPAuthenticatorData {
  const addedOn = formatDatetime(locale, authenticator.createdAt) ?? "";
  const verificationMethod = getOobOtpVerificationMethod(authenticator);
  const iconName = oobOtpVerificationMethodIconName[verificationMethod];
  const label = getOobOtpAuthenticatorLabel(authenticator, verificationMethod);

  return {
    id: authenticator.id,
    kind: authenticator.kind,
    isDefault: authenticator.isDefault,
    iconName,
    label,
    addedOn,
  };
}

function constructAuthenticatorLists(
  authenticators: Authenticator[],
  kind: AuthenticatorKind,
  locale: string
) {
  const passwordAuthenticatorList: PasswordAuthenticatorData[] = [];
  const oobOtpAuthenticatorList: OOBOTPAuthenticatorData[] = [];
  const totpAuthenticatorList: TOTPAuthenticatorData[] = [];

  const filteredAuthenticators = authenticators.filter((a) => a.kind === kind);

  for (const authenticator of filteredAuthenticators) {
    switch (authenticator.type) {
      case "PASSWORD":
        passwordAuthenticatorList.push(
          constructPasswordAuthenticatorData(authenticator, locale)
        );
        break;
      case "OOB_OTP":
        oobOtpAuthenticatorList.push(
          constructOobOtpAuthenticatorData(authenticator, locale)
        );
        break;
      case "TOTP":
        if (kind === "PRIMARY") {
          break;
        }
        totpAuthenticatorList.push(
          constructTotpAuthenticatorData(authenticator, locale)
        );
        break;
      default:
        break;
    }
  }

  return kind === "PRIMARY"
    ? {
        password: passwordAuthenticatorList,
        oobOtp: oobOtpAuthenticatorList,
        hasVisibleList: [
          passwordAuthenticatorList,
          oobOtpAuthenticatorList,
        ].some((list) => list.length > 0),
      }
    : {
        password: passwordAuthenticatorList,
        oobOtp: oobOtpAuthenticatorList,
        totp: totpAuthenticatorList,
        hasVisibleList: [
          passwordAuthenticatorList,
          oobOtpAuthenticatorList,
          totpAuthenticatorList,
        ].some((list) => list.length > 0),
      };
}

const RemoveConfirmationDialog: React.FC<RemoveConfirmationDialogProps> = function RemoveConfirmationDialog(
  props: RemoveConfirmationDialogProps
) {
  const {
    visible,
    deleteAuthenticator,
    deletingAuthenticator,
    authenticatorID,
    authenticatorName,
    onDismiss,
  } = props;

  const { renderToString } = useContext(Context);

  const onConfirmClicked = useCallback(() => {
    deleteAuthenticator(authenticatorID!);
  }, [deleteAuthenticator, authenticatorID]);

  const dialogMessage = useMemo(() => {
    return renderToString(
      "UserDetails.account-security.remove-confirm-dialog.message",
      { authenticatorName: authenticatorName ?? "" }
    );
  }, [renderToString, authenticatorName]);

  return (
    <Dialog
      hidden={!visible}
      title={
        <FormattedMessage id="UserDetails.account-security.remove-confirm-dialog.title" />
      }
      subText={dialogMessage}
      onDismiss={onDismiss}
    >
      <DialogFooter>
        <ButtonWithLoading
          onClick={onConfirmClicked}
          labelId="confirm"
          loading={deletingAuthenticator}
        />
        <DefaultButton onClick={onDismiss}>
          <FormattedMessage id="cancel" />
        </DefaultButton>
      </DialogFooter>
    </Dialog>
  );
};

const ErrorDialog: React.FC<ErrorDialogProps> = function ErrorDialog(
  props: ErrorDialogProps
) {
  const { visible, errorMessage, onDismiss } = props;

  return (
    <Dialog
      hidden={!visible}
      title={<FormattedMessage id="error" />}
      subText={errorMessage}
      onDismiss={onDismiss}
    >
      <DialogFooter>
        <PrimaryButton onClick={onDismiss}>
          <FormattedMessage id="ok" />
        </PrimaryButton>
      </DialogFooter>
    </Dialog>
  );
};

const PasswordAuthenticatorCell: React.FC<PasswordAuthenticatorCellProps> = function PasswordAuthenticatorCell(
  props: PasswordAuthenticatorCellProps
) {
  const { id, kind, lastUpdated, showConfirmationDialog } = props;
  const navigate = useNavigate();
  const { renderToString } = useContext(Context);

  const labelId = getLocaleKeyWithAuthenticatorType("PASSWORD", kind);

  const onResetPasswordClicked = useCallback(() => {
    navigate("./reset-password");
  }, [navigate]);

  const onRemoveClicked = useCallback(() => {
    showConfirmationDialog(id, renderToString(labelId!));
  }, [labelId, id, renderToString, showConfirmationDialog]);

  return (
    <ListCellLayout className={cn(styles.cell, styles.passwordCell)}>
      <Text className={cn(styles.cellLabel, styles.passwordCellLabel)}>
        <FormattedMessage id={labelId!} />
      </Text>
      <Text className={cn(styles.cellDesc, styles.passwordCellDesc)}>
        <FormattedMessage
          id="UserDetails.account-security.last-updated"
          values={{ datetime: lastUpdated }}
        />
      </Text>
      {kind === "PRIMARY" && (
        <PrimaryButton
          className={cn(styles.button, styles.resetPasswordButton)}
          onClick={onResetPasswordClicked}
        >
          <FormattedMessage id="UserDetails.account-security.reset-password" />
        </PrimaryButton>
      )}
      {kind === "SECONDARY" && (
        <DefaultButton
          className={cn(
            styles.button,
            styles.removeButton,
            styles.removePasswordButton
          )}
          onClick={onRemoveClicked}
          theme={destructiveTheme}
        >
          <FormattedMessage id="remove" />
        </DefaultButton>
      )}
    </ListCellLayout>
  );
};

const TOTPAuthenticatorCell: React.FC<TOTPAuthenticatorCellProps> = function TOTPAuthenticatorCell(
  props: TOTPAuthenticatorCellProps
) {
  const { id, kind, label, addedOn, showConfirmationDialog } = props;

  const onRemoveClicked = useCallback(() => {
    showConfirmationDialog(id, label);
  }, [id, label, showConfirmationDialog]);

  return (
    <ListCellLayout className={cn(styles.cell, styles.totpCell)}>
      <Text className={cn(styles.cellLabel, styles.totpCellLabel)}>
        {label}
      </Text>
      <Text className={cn(styles.cellDesc, styles.totpCellDesc)}>
        <FormattedMessage
          id="UserDetails.account-security.added-on"
          values={{ datetime: addedOn }}
        />
      </Text>
      {kind === "SECONDARY" && (
        <DefaultButton
          className={cn(
            styles.button,
            styles.removeButton,
            styles.totpRemoveButton
          )}
          onClick={onRemoveClicked}
          theme={destructiveTheme}
        >
          <FormattedMessage id="remove" />
        </DefaultButton>
      )}
    </ListCellLayout>
  );
};

const OOBOTPAuthenticatorCell: React.FC<OOBOTPAuthenticatorCellProps> = function (
  props: OOBOTPAuthenticatorCellProps
) {
  const { id, label, iconName, kind, addedOn, showConfirmationDialog } = props;

  const onRemoveClicked = useCallback(() => {
    showConfirmationDialog(id, label);
  }, [id, label, showConfirmationDialog]);

  return (
    <ListCellLayout className={cn(styles.cell, styles.oobOtpCell)}>
      <Icon className={styles.oobOtpCellIcon} iconName={iconName} />
      <Text className={cn(styles.cellLabel, styles.oobOtpCellLabel)}>
        {label}
      </Text>
      <Text className={cn(styles.cellDesc, styles.oobOtpCellAddedOn)}>
        <FormattedMessage
          id="UserDetails.account-security.added-on"
          values={{ datetime: addedOn }}
        />
      </Text>

      {kind === "SECONDARY" && (
        <DefaultButton
          className={cn(
            styles.button,
            styles.removeButton,
            styles.oobOtpRemoveButton
          )}
          onClick={onRemoveClicked}
          theme={destructiveTheme}
        >
          <FormattedMessage id="remove" />
        </DefaultButton>
      )}
    </ListCellLayout>
  );
};

const UserDetailsAccountSecurity: React.FC<UserDetailsAccountSecurityProps> = function UserDetailsAccountSecurity(
  props: UserDetailsAccountSecurityProps
) {
  const { authenticators } = props;
  const { locale } = useContext(Context);

  const {
    deleteAuthenticator,
    loading: deletingAuthenticator,
    error: deleteAuthenticatorError,
  } = useDeleteAuthenticatorMutation();

  const [
    confirmationDialogData,
    setConfirmationDialogData,
  ] = useState<RemoveConfirmationDialogData | null>(null);
  const [
    errorDialogData,
    setErrorDialogData,
  ] = useState<ErrorDialogData | null>(null);

  const [violations, setViolations] = useState<Violation[]>([]);
  const [unhandledViolations, setUnhandledViolations] = useState<Violation[]>(
    []
  );

  const primaryAuthenticatorLists = useMemo(() => {
    return constructAuthenticatorLists(authenticators, "PRIMARY", locale);
  }, [locale, authenticators]);

  const secondaryAuthenticatorLists = useMemo(() => {
    return constructAuthenticatorLists(authenticators, "SECONDARY", locale);
  }, [locale, authenticators]);

  const showConfirmationDialog = useCallback(
    (authenticatorID: string, authenticatorName: string) => {
      setConfirmationDialogData({
        authenticatorID,
        authenticatorName,
      });
    },
    []
  );

  const dismissConfirmationDialog = useCallback(() => {
    setConfirmationDialogData(null);
  }, []);

  const onRenderPasswordAuthenticatorDetailCell = useCallback(
    (item?: PasswordAuthenticatorData, _index?: number): React.ReactNode => {
      if (item == null) {
        return null;
      }
      return (
        <PasswordAuthenticatorCell
          {...item}
          showConfirmationDialog={showConfirmationDialog}
        />
      );
    },
    [showConfirmationDialog]
  );

  const onRenderOobOtpAuthenticatorDetailCell = useCallback(
    (item?: OOBOTPAuthenticatorData, _index?: number): React.ReactNode => {
      if (item == null) {
        return null;
      }
      return (
        <OOBOTPAuthenticatorCell
          {...item}
          showConfirmationDialog={showConfirmationDialog}
        />
      );
    },
    [showConfirmationDialog]
  );

  const onRenderTotpAuthenticatorDetailCell = useCallback(
    (item?: TOTPAuthenticatorData, _index?: number): React.ReactNode => {
      if (item == null) {
        return null;
      }
      return (
        <TOTPAuthenticatorCell
          {...item}
          showConfirmationDialog={showConfirmationDialog}
        />
      );
    },
    [showConfirmationDialog]
  );

  const onConfirmDeleteAuthenticator = useCallback(
    (authenticatorID) => {
      deleteAuthenticator(authenticatorID)
        .then((success) => {
          if (!success) {
            throw new Error();
          }
        })
        .catch((err) => {
          const newViolations = parseError(err);
          setViolations(newViolations);
        })
        .finally(() => {
          dismissConfirmationDialog();
        });
    },
    [deleteAuthenticator, dismissConfirmationDialog]
  );

  const errorMessage = useMemo(() => {
    const errorDialogErrorMessages: string[] = [];
    const unknownViolations: Violation[] = [];

    for (const violation of violations) {
      unknownViolations.push(violation);
    }

    setUnhandledViolations(unknownViolations);

    return {
      errorDialog: defaultFormatErrorMessageList(errorDialogErrorMessages),
    };
  }, [violations]);

  useEffect(() => {
    if (errorMessage.errorDialog == null) {
      setErrorDialogData(null);
    } else {
      setErrorDialogData({ errorMessage: errorMessage.errorDialog });
    }
  }, [errorMessage.errorDialog]);

  const dismissErrorDialog = useCallback(() => {
    setErrorDialogData(null);
  }, []);

  return (
    <div className={styles.root}>
      {unhandledViolations.length > 0 && (
        <ShowError error={deleteAuthenticatorError} />
      )}
      <RemoveConfirmationDialog
        visible={confirmationDialogData != null}
        authenticatorID={confirmationDialogData?.authenticatorID}
        authenticatorName={confirmationDialogData?.authenticatorName}
        onDismiss={dismissConfirmationDialog}
        deleteAuthenticator={onConfirmDeleteAuthenticator}
        deletingAuthenticator={deletingAuthenticator}
      />
      <ErrorDialog
        visible={errorDialogData != null}
        errorMessage={errorDialogData?.errorMessage}
        onDismiss={dismissErrorDialog}
      />
      {primaryAuthenticatorLists.hasVisibleList && (
        <div className={styles.authenticatorContainer}>
          <Text
            as="h2"
            className={cn(styles.header, styles.authenticatorKindHeader)}
          >
            <FormattedMessage id="UserDetails.account-security.primary" />
          </Text>
          {primaryAuthenticatorLists.password.length > 0 && (
            <List
              className={styles.list}
              items={primaryAuthenticatorLists.password}
              onRenderCell={onRenderPasswordAuthenticatorDetailCell}
            />
          )}
          {primaryAuthenticatorLists.oobOtp.length > 0 && (
            <>
              <Text
                as="h3"
                className={cn(styles.header, styles.authenticatorTypeHeader)}
              >
                <FormattedMessage id="UserDetails.account-security.primary.oob-otp" />
              </Text>
              <List
                className={cn(styles.list, styles.oobOtpList)}
                items={primaryAuthenticatorLists.oobOtp}
                onRenderCell={onRenderOobOtpAuthenticatorDetailCell}
              />
            </>
          )}
        </div>
      )}
      {secondaryAuthenticatorLists.hasVisibleList && (
        <div className={styles.authenticatorContainer}>
          <Text
            as="h2"
            className={cn(styles.header, styles.authenticatorKindHeader)}
          >
            <FormattedMessage id="UserDetails.account-security.secondary" />
          </Text>
          {secondaryAuthenticatorLists.totp != null &&
            secondaryAuthenticatorLists.totp.length > 0 && (
              <>
                <Text
                  as="h3"
                  className={cn(styles.header, styles.authenticatorTypeHeader)}
                >
                  <FormattedMessage id="UserDetails.account-security.secondary.totp" />
                </Text>
                <List
                  className={cn(styles.list, styles.totpList)}
                  items={secondaryAuthenticatorLists.totp}
                  onRenderCell={onRenderTotpAuthenticatorDetailCell}
                />
              </>
            )}
          {secondaryAuthenticatorLists.oobOtp.length > 0 && (
            <>
              <Text
                as="h3"
                className={cn(styles.header, styles.authenticatorTypeHeader)}
              >
                <FormattedMessage id="UserDetails.account-security.secondary.oob-otp" />
              </Text>
              <List
                className={cn(styles.list, styles.oobOtpList)}
                items={secondaryAuthenticatorLists.oobOtp}
                onRenderCell={onRenderOobOtpAuthenticatorDetailCell}
              />
            </>
          )}
          {secondaryAuthenticatorLists.password.length > 0 && (
            <List
              className={cn(styles.list, styles.passwordList)}
              items={secondaryAuthenticatorLists.password}
              onRenderCell={onRenderPasswordAuthenticatorDetailCell}
            />
          )}
        </div>
      )}
    </div>
  );
};

export default UserDetailsAccountSecurity;

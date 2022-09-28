import React, { ReactNode, ReactElement } from "react";
import {
  CompoundButton,
  IButtonStyles,
  IButtonProps,
  useTheme,
} from "@fluentui/react";

export interface ChoiceButtonProps {
  className?: string;
  checked?: IButtonProps["checked"];
  disabled?: IButtonProps["disabled"];
  text?: ReactNode;
  secondaryText?: ReactNode;
  onClick?: IButtonProps["onClick"];
}

export default function ChoiceButton(props: ChoiceButtonProps): ReactElement {
  const originalTheme = useTheme();
  const styles: IButtonStyles = {
    root: {
      maxWidth: "auto",
    },
    rootChecked: {
      borderColor: originalTheme.palette.themePrimary,
      backgroundColor: originalTheme.semanticColors.buttonBackground,
    },
    description: {
      color: "inherit",
    },
    // When ChoiceButton is taller than its intrinsic height,
    // make sure the content is still center aligned vertically.
    flexContainer: {
      alignItems: "center",
    },
  };
  // @ts-expect-error
  return <CompoundButton {...props} toggle={true} styles={styles} />;
}

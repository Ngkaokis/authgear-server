import React, { useContext } from "react";
import { Spinner } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import styles from "./ShowLoading.module.css";
import cn from "classnames";

interface ShowLoadingProps {
  label?: string;
}

// ShowLoading show a 100% width and 100% height spinner.
// For better UX, please use Shimmer instead.
const ShowLoading: React.FC<ShowLoadingProps> = function ShowLoading({
  label,
}) {
  const { renderToString } = useContext(Context);

  return (
    <div className={cn(styles.loading, "mobile:fixed mobile:align-middle")}>
      <Spinner label={label ?? renderToString("loading")} />
    </div>
  );
};

export default ShowLoading;

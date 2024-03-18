import {
  Dropdown,
  IDropdownOption,
  IDropdownProps,
  SearchBox,
  Spinner,
  SpinnerSize,
} from "@fluentui/react";
import React, { useCallback, useContext, useMemo } from "react";
import { Context as MessageContext } from "@oursky/react-messageformat";

import styles from "./SearchableDropdown.module.css";

interface SearchableDropdownProps
  extends Omit<
    IDropdownProps,
    | "defaultSelectedKey"
    | "selectedKey"
    | "defaultSelectedKeys"
    | "selectedKeys"
  > {
  isLoadingOptions?: boolean;
  onSearchValueChange?: (value: string) => void;
  searchValue?: string;
  searchPlaceholder?: string;
  selectedItem?: IDropdownOption | null;
  selectedItems?: IDropdownOption[];
}

function SearchableDropdownSearchBox(props: {
  onValueChange: ((value: string) => void) | undefined;
  value: string | undefined;
  placeholder: string | undefined;
}) {
  const { renderToString } = useContext(MessageContext);
  const { value, onValueChange, placeholder } = props;

  const onChange = useCallback(
    (e?: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      if (e == null) {
        return;
      }
      const value = e.currentTarget.value;
      onValueChange?.(value);
    },
    [onValueChange]
  );

  const onClear = useCallback(() => {
    onValueChange?.("");
  }, [onValueChange]);

  return (
    <SearchBox
      placeholder={placeholder ?? renderToString("search")}
      underlined={true}
      value={value}
      onChange={onChange}
      onClear={onClear}
    />
  );
}

const EMPTY_CALLOUT_PROPS: IDropdownProps["calloutProps"] = {};

export const SearchableDropdown: React.VFC<SearchableDropdownProps> =
  function SearchableDropdown(props) {
    const {
      options,
      isLoadingOptions,
      onSearchValueChange,
      searchValue,
      searchPlaceholder,
      calloutProps = EMPTY_CALLOUT_PROPS,
      selectedItem,
      selectedItems,
      ...restProps
    } = props;

    const onRenderList = useCallback<
      NonNullable<IDropdownProps["onRenderList"]>
    >(
      (props?, defaultRenderer?) => {
        if (defaultRenderer == null) {
          return null;
        }

        return (
          <>
            <div className={styles.searchBoxRow}>
              <SearchableDropdownSearchBox
                onValueChange={onSearchValueChange}
                value={searchValue}
                placeholder={searchPlaceholder}
              />
            </div>
            {isLoadingOptions ? (
              <div className={styles.optionsLoadingRow}>
                <Spinner size={SpinnerSize.xSmall} />
              </div>
            ) : (
              defaultRenderer(props)
            )}
          </>
        );
      },
      [isLoadingOptions, onSearchValueChange, searchPlaceholder, searchValue]
    );

    const combinedOptions = useMemo(() => {
      const providedOptionKeys = new Set(options.map((o) => o.key));

      // Include all selected items as hidden options, if they are not in `options`.
      // This is needed for the dropdown to display selected options correctly.
      const hiddenOptions: IDropdownOption[] = [];
      if (selectedItem && !providedOptionKeys.has(selectedItem.key)) {
        hiddenOptions.push({ ...selectedItem, hidden: true });
      }
      if (selectedItems) {
        for (const item of selectedItems) {
          if (!providedOptionKeys.has(item.key)) {
            hiddenOptions.push({ ...item, hidden: true });
          }
        }
      }
      return options.concat(hiddenOptions);
    }, [options, selectedItem, selectedItems]);

    const selectedKey = useMemo(() => {
      return selectedItem?.key;
    }, [selectedItem]);

    const selectedKeys = useMemo(() => {
      return selectedItems?.map((item) => item.key);
    }, [selectedItems]);

    return (
      <Dropdown
        options={combinedOptions}
        onRenderList={onRenderList}
        {...restProps}
        calloutProps={{
          calloutMaxHeight: 264,
          calloutMinWidth: 200,
          alignTargetEdge: true,
          ...calloutProps,
        }}
        selectedKey={selectedKey}
        selectedKeys={selectedKeys as IDropdownProps["selectedKeys"]}
      />
    );
  };

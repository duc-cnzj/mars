import classNames from "classnames";
import React from "react";

interface IconFontProps extends React.ComponentPropsWithoutRef<"svg"> {
  name: string;
  className?: string;
}

const IconFont: React.FC<IconFontProps> = ({ name, className, ...props }) => {
  return (
    <span className="anticon anticon-money-collect ant-menu-item-icon">
      <svg
        className={classNames("icon", className)}
        aria-hidden="true"
        {...props}
      >
        <use fill="currentColor" xlinkHref={name} />
      </svg>
    </span>
  );
};

export default IconFont;

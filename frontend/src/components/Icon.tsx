import React from "react";

interface IconFontProps extends React.ComponentPropsWithoutRef<"svg"> {
  name: string;
}

const IconFont: React.FC<IconFontProps> = ({ name, ...props }) => {
  return (
    <span className="anticon anticon-money-collect ant-menu-item-icon">
      <svg className="icon" aria-hidden="true" {...props}>
        <use fill="currentColor" xlinkHref={name} />
      </svg>
    </span>
  );
};

export default IconFont;

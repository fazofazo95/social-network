import Image from "next/image";

const Input = ({ 
  label, 
  icon, 
  required = false, 
  optional = false,
  id, 
  type = "text", 
  placeholder,
  className = "",
  ...props 
}) => {
  return (
    <div className={`relative ${className}`}>
      <label className="label-custom" htmlFor={id}>
        {label}{" "}
        {required && <span className="text-red-500">*</span>}
        {optional && <span className="text-gray-500">(Optional)</span>}
      </label>

      {icon && (
        <Image
          src={icon}
          alt={`${label} Icon`}
          width={20}
          height={20}
          className="absolute left-2 top-3"
        />
      )}
      
      <input
        className="input focus:outline-none"
        id={id}
        type={type}
        placeholder={placeholder}
        {...props}
      />
    </div>
  );
};

export default Input;

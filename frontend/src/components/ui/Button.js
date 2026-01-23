const Button = ({ 
  children, 
  type = "button", 
  variant = "primary",
  className = "",
  ...props 
}) => {
  const baseStyles = "font-bold py-2 px-25 w-full rounded focus:outline-none focus:shadow-outline";
  
  const variants = {
    primary: "bg-blue-500 hover:bg-blue-700 text-white",
    secondary: "bg-gray-500 hover:bg-gray-700 text-white",
    danger: "bg-red-500 hover:bg-red-700 text-white",
  };

  return (
    <button
      className={`${baseStyles} ${variants[variant]} ${className}`}
      type={type}
      {...props}
    >
      {children}
    </button>
  );
};

export default Button;

import style from "./button.module.scss"

interface ButtonProps extends React.DetailedHTMLProps<React.ButtonHTMLAttributes<HTMLButtonElement>, HTMLButtonElement> {
    text: string
}

function Button(props: ButtonProps) {
    const { text } = props;

    return (
        <button className={style.button} {...props}>
            <span>{text}</span>
        </button>
    )
}

export default Button
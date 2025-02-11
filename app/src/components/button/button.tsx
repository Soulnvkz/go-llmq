import style from "./button.module.scss"

interface ButtonProps {
    text: string
}

function Button({text}: ButtonProps) {
    return (
        <button className={style.button}>
            <span>{text}</span>
        </button>
    )
}

export default Button
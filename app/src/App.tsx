import './App.css'
import Button from './components/button/button'
import Message from './components/message/message'
import MessagesContainer from './components/message/messagesContainer'

function App() {
  return (
    <div className="chat-container">
      <div className="chat-header">
        <h1>Text Chat #1</h1>
      </div>
      <MessagesContainer>
        <Message text="Hello! How can I help you today?"
          isUser={false}
        />
        <Message text="dunno know"
          isUser={true}
        />
      </MessagesContainer>

      <div className="input-container">
        <input type="text" className="message-input" placeholder="Type your message..." />
        <Button text="Send" />
      </div>
    </div>)
}

export default App

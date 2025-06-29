// console.log("main.js is running!");

// Getting all elements from the page
const chatContainer = document.getElementById('chat-container');
const inputForm = document.getElementById('input-form');
const userInput = document.getElementById('user-input');
const imageUpload = document.getElementById('image-upload');
const imagePreviewContainer = document.getElementById('image-preview');
const previewImage = document.getElementById('preview-image');

const botName = "TeacherAI"
const userName = "User"


// Appending a new message to the chat container
// Helper function
function addMessage(sender, text, imageSrc = null, materials = null) {
    const messageDiv = document.createElement('div');
    var senderType;
    let renderedText = text;
    
    // Switch the type depending on the sender
    if (sender == botName){    
        senderType = "bot" 
        renderedText = marked.parse(text);// Use the Markdown library to convert the bot's text to HTML

    } else if (sender == userName){
        senderType = "user"
    }

    // console.log("sender "+ sender)
    // console.log("sendertype "+ senderType)

    messageDiv.className = `message ${senderType}`; // Uses CSS classes: 'user' or 'bot'
    messageDiv.innerHTML = `<strong>${sender}:</strong> ${renderedText}`; // Adding HTML content to the message, with the name instead of the type.

    
    // If an image was provided, add it to the message bubble
    if (imageSrc) {
        const img = document.createElement('img');
        img.src = imageSrc;
        img.style.maxWidth = '200px';
        img.style.maxHeight = '200px';
        img.style.marginTop = '10px';
        messageDiv.appendChild(img);
    }

    // If material recommendations were provided, render them.
    // Showing the DUMMY article as links and videos as an embedded HTML
    if (materials) {
        if (materials.articles && materials.articles.length > 0) {
            messageDiv.innerHTML += '<p><strong>Recommended Articles:</strong></p>';
            materials.articles.forEach(item => {
                messageDiv.innerHTML += `<a href="${item.url}" target="_blank">${item.title}</a><br>`;
            });
        }
        if (materials.videos && materials.videos.length > 0) {
            messageDiv.innerHTML += '<p><strong>Recommended Videos:</strong></p>';
            materials.videos.forEach(item => {
                // Check if the URL is a YouTube video link
                if (item.url.includes("youtube.com/watch") || item.url.includes("youtu.be")) {
                    // Extract the video ID from the URL
                    const videoID = item.url.split("v=")[1] || item.url.split("/")[3];
                    const embedURL = `https://www.youtube.com/embed/${videoID}`;
                    
                    // Embed the video using an iframe.
                    messageDiv.innerHTML += `<div style="margin-top: 10px;"><iframe width="70%" height="20%" src="${embedURL}" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe></div>`;
                } else {
                    // If it's not a YouTube video, just show a regular link
                    messageDiv.innerHTML += `<a href="${item.url}" target="_blank">${item.title}</a><br>`;
                }
            });
        }
    }


    // Add the new message to the chat display
    chatContainer.appendChild(messageDiv);
    
    // Scroll to the bottom of the chat to see the latest message
    chatContainer.scrollTop = chatContainer.scrollHeight;
}

// Add an initial message 
addMessage(botName, 'Welcome! I am your learning assistant AI. I can help you explore a topic or dive deep into the details. Ask me questions, upload images for context, or even ask me to find videos on a topic you want to learn about!');


// Listening when there's an image being uploaded
imageUpload.addEventListener('change', (event) => {
    const file = event.target.files[0];
    if (file) {
        const reader = new FileReader();
        reader.onload = (e) => {
            previewImage.src = e.target.result; // Set the image preview's source
            imagePreviewContainer.style.display = 'block'; // Make the preview visible
        };
        reader.readAsDataURL(file); // Read the file as a data URL (Base64 string)
    } else {
        imagePreviewContainer.style.display = 'none';
        previewImage.src = '';
    }
});

// Handles the submit event
inputForm.addEventListener('submit', async (e) => {
    e.preventDefault(); // Not refresing
    
    const message = userInput.value;
    const imageFile = imageUpload.files[0];

    if (message.trim() === '' && !imageFile) {
        return; // Don't send if both input and image are empty
    }

    // Display the user's message and/or image in the chat immediately
    const imageUrl = imageFile ? URL.createObjectURL(imageFile) : null;
    addMessage(userName, message, imageUrl);

    userInput.value = ''; // Clear the text input
    imageUpload.value = ''; // Clear the file input
    imagePreviewContainer.style.display = 'none'; // Hide the preview

    // Create a FormData object to bundle the text and the file for sending
    const formData = new FormData();
    formData.append('message', message);
    if (imageFile) {
        formData.append('image', imageFile);
    }
    
    // Show a "loading" message from the bot while we wait for the API
    addMessage(botName, 'Thinking...');
    
    try {
        // Use the fetch() API to send a POST request to Go backend
        const response = await fetch('/chat', {
            method: 'POST',
            body: formData, 
        });
        const data = await response.json();
        
        // Remove the "Thinking..." message when done parsing
        const lastMessage = chatContainer.lastElementChild;
        if (lastMessage && lastMessage.textContent.includes('Thinking...')) {
            chatContainer.removeChild(lastMessage);
        }

        // Check if the response contains material recommendations
        if (data.articles || data.videos) {
            // If it does, render the Dummy data materials only
            addMessage(botName, data.bot_response, null, { articles: data.articles, videos: data.videos });
        } else {
            // Otherwise, just display the regular text response
            addMessage(botName, data.bot_response);
        }

    } catch (error) {
        console.error('Error:', error);
        const lastMessage = chatContainer.lastElementChild;
        if (lastMessage && lastMessage.textContent.includes('Thinking...')) {
            chatContainer.removeChild(lastMessage);
        }
        addMessage(botName, 'The server ran into an error.');
    }
});



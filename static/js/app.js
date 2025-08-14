// Poker Planning functionality

// Voting card selection
document.addEventListener('DOMContentLoaded', function() {
    // Handle voting card selection
    document.addEventListener('click', function(e) {
        if (e.target.matches('.voting-card') || e.target.closest('.voting-card')) {
            const card = e.target.matches('.voting-card') ? e.target : e.target.closest('.voting-card');
            
            // Remove selected class from all cards
            document.querySelectorAll('.voting-card').forEach(c => c.classList.remove('selected'));
            
            // Add selected class to clicked card
            card.classList.add('selected');
        }
    });
});

// Copy session URL to clipboard
function copySessionUrl() {
    const url = window.location.href;
    navigator.clipboard.writeText(url).then(() => {
        // Show success feedback
        const feedback = document.createElement('div');
        feedback.textContent = 'Session URL copied to clipboard!';
        feedback.className = 'fixed top-4 right-4 bg-green-500 text-white px-4 py-2 rounded shadow-lg z-50';
        document.body.appendChild(feedback);
        
        setTimeout(() => {
            feedback.remove();
        }, 2000);
    });
}

// Keyboard shortcuts
document.addEventListener('keydown', function(e) {
    // Number keys 1-9 for voting cards
    if (e.key >= '1' && e.key <= '9' && !e.ctrlKey && !e.altKey && !e.metaKey) {
        const cardIndex = parseInt(e.key) - 1;
        const cards = document.querySelectorAll('.voting-card:not([disabled])');
        if (cards[cardIndex]) {
            cards[cardIndex].click();
        }
    }
    
    // Space to reveal votes (owner only)
    if (e.key === ' ' && e.target.tagName !== 'INPUT' && e.target.tagName !== 'TEXTAREA') {
        e.preventDefault();
        const revealBtn = document.querySelector('[hx-post*="end-voting"], [hx-post*="start-voting"]');
        if (revealBtn && !revealBtn.disabled) {
            revealBtn.click();
        }
    }
});
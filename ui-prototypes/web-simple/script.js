class ConsensusUI {
    constructor() {
        this.initializeElements();
        this.attachEventListeners();
        this.results = [];
    }

    initializeElements() {
        this.form = document.getElementById('consensus-form');
        this.promptInput = document.getElementById('prompt');
        this.masterPromptToggle = document.getElementById('use-master-prompt');
        this.masterProviderSection = document.getElementById('master-provider-section');
        this.masterProviderSelect = document.getElementById('master-provider');
        this.responseProviderCheckboxes = document.querySelectorAll('.provider-checkboxes input[type="checkbox"]');
        this.emailInput = document.getElementById('email-recipients');
        this.submitBtn = document.getElementById('submit-btn');
        this.resultsSection = document.getElementById('results');
        this.resultsContainer = document.getElementById('results-container');
    }

    attachEventListeners() {
        this.form.addEventListener('submit', (e) => this.handleSubmit(e));
        this.masterPromptToggle.addEventListener('change', () => this.toggleMasterPrompt());
        
        // Validate at least one provider is selected
        this.responseProviderCheckboxes.forEach(checkbox => {
            checkbox.addEventListener('change', () => this.validateProviders());
        });
    }

    toggleMasterPrompt() {
        const isEnabled = this.masterPromptToggle.checked;
        this.masterProviderSection.style.display = isEnabled ? 'block' : 'none';
    }

    validateProviders() {
        const selectedProviders = Array.from(this.responseProviderCheckboxes)
            .filter(cb => cb.checked);
        
        if (selectedProviders.length === 0) {
            this.submitBtn.disabled = true;
            this.showError('Please select at least one response provider');
        } else {
            this.submitBtn.disabled = false;
            this.clearError();
        }
    }

    async handleSubmit(e) {
        e.preventDefault();
        
        const formData = this.collectFormData();
        if (!this.validateForm(formData)) return;

        this.setLoadingState(true);
        this.showResults();
        
        try {
            await this.sendRequest(formData);
        } catch (error) {
            this.showError(`Request failed: ${error.message}`);
        } finally {
            this.setLoadingState(false);
        }
    }

    collectFormData() {
        const selectedProviders = Array.from(this.responseProviderCheckboxes)
            .filter(cb => cb.checked)
            .map(cb => cb.value);

        const emails = this.emailInput.value
            .split(',')
            .map(email => email.trim())
            .filter(email => email);

        return {
            prompt: this.promptInput.value.trim(),
            useMasterPrompt: this.masterPromptToggle.checked,
            masterProvider: this.masterProviderSelect.value,
            responseProviders: selectedProviders,
            emailRecipients: emails
        };
    }

    validateForm(data) {
        if (!data.prompt) {
            this.showError('Please enter a prompt');
            return false;
        }

        if (data.responseProviders.length === 0) {
            this.showError('Please select at least one response provider');
            return false;
        }

        // Validate email format if provided
        if (data.emailRecipients.length > 0) {
            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            const invalidEmails = data.emailRecipients.filter(email => !emailRegex.test(email));
            if (invalidEmails.length > 0) {
                this.showError(`Invalid email addresses: ${invalidEmails.join(', ')}`);
                return false;
            }
        }

        return true;
    }

    async sendRequest(data) {
        // This would connect to your Go backend API
        // For now, we'll simulate the behavior
        
        const sessionId = this.generateSessionId();
        this.clearResults();
        
        // Simulate master prompt optimization if enabled
        if (data.useMasterPrompt) {
            this.addResult({
                provider: data.masterProvider,
                type: 'master-prompt',
                status: 'pending',
                content: 'Optimizing your prompt...'
            });

            await this.simulateDelay(1000);
            
            this.updateResult(data.masterProvider, 'master-prompt', {
                status: 'complete',
                content: `Optimized prompt created using ${data.masterProvider.toUpperCase()}:\n\n"${data.prompt}" -> "Create a comprehensive and detailed response to: ${data.prompt}. Ensure the response is well-structured, informative, and addresses all aspects of the request."`
            });
        }

        // Simulate responses from each provider
        const promises = data.responseProviders.map(async (provider) => {
            this.addResult({
                provider: provider,
                type: 'response',
                status: 'pending',
                content: `Waiting for ${provider.toUpperCase()} response...`
            });

            // Simulate varying response times
            const delay = Math.random() * 3000 + 1000;
            await this.simulateDelay(delay);

            const mockResponse = this.generateMockResponse(provider, data.prompt);
            
            this.updateResult(provider, 'response', {
                status: 'complete',
                content: mockResponse
            });
        });

        await Promise.all(promises);

        // Simulate email notification
        if (data.emailRecipients.length > 0) {
            await this.simulateDelay(500);
            this.showSuccessMessage(`Email notifications sent to: ${data.emailRecipients.join(', ')}`);
        }
    }

    addResult(result) {
        const resultElement = document.createElement('div');
        resultElement.className = 'result-card';
        resultElement.dataset.provider = result.provider;
        resultElement.dataset.type = result.type;
        
        resultElement.innerHTML = `
            <div class="result-header">
                <span class="provider-name">${result.provider.toUpperCase()} ${result.type === 'master-prompt' ? '(Master Prompt)' : ''}</span>
                <span class="status-indicator status-${result.status}">${result.status.charAt(0).toUpperCase() + result.status.slice(1)}</span>
            </div>
            <div class="result-content">${result.content}</div>
        `;
        
        this.resultsContainer.appendChild(resultElement);
    }

    updateResult(provider, type, updates) {
        const resultElement = document.querySelector(`[data-provider="${provider}"][data-type="${type}"]`);
        if (!resultElement) return;

        if (updates.status) {
            const statusElement = resultElement.querySelector('.status-indicator');
            statusElement.textContent = updates.status.charAt(0).toUpperCase() + updates.status.slice(1);
            statusElement.className = `status-indicator status-${updates.status}`;
        }

        if (updates.content) {
            const contentElement = resultElement.querySelector('.result-content');
            contentElement.textContent = updates.content;
        }
    }

    generateMockResponse(provider, prompt) {
        const responses = {
            openai: `GPT-4 Response to "${prompt}":

This is a comprehensive analysis of your request. I've considered multiple perspectives and approaches to provide you with a well-rounded response.

Key points:
1. Understanding the context and requirements
2. Analyzing potential solutions and approaches
3. Providing actionable recommendations
4. Considering potential challenges and mitigation strategies

The response demonstrates OpenAI's focus on structured, analytical thinking with practical applications.`,

            anthropic: `Claude's Response to "${prompt}":

I appreciate the thoughtful nature of your question. Let me break this down systematically:

First, let's consider the fundamental aspects:
- The core principles involved
- The broader implications and context
- Multiple stakeholder perspectives

My analysis suggests several key considerations:
1. Immediate practical steps
2. Long-term strategic thinking
3. Ethical implications and best practices

This response reflects Anthropic's emphasis on helpful, harmless, and honest AI assistance.`,

            gemini: `Gemini's Response to "${prompt}":

Your question touches on several interesting dimensions that I'd like to explore:

🔍 Analysis:
- Current state assessment
- Opportunity identification
- Risk evaluation

💡 Insights:
- Pattern recognition from similar scenarios
- Creative approaches and alternatives
- Data-driven recommendations

🎯 Action Items:
- Prioritized next steps
- Success metrics and milestones
- Continuous improvement strategies

This demonstrates Google's focus on comprehensive, multi-modal analysis with actionable outcomes.`
        };

        return responses[provider] || `Mock response from ${provider.toUpperCase()} for: "${prompt}"`;
    }

    clearResults() {
        this.resultsContainer.innerHTML = '';
    }

    showResults() {
        this.resultsSection.style.display = 'block';
    }

    setLoadingState(isLoading) {
        this.submitBtn.disabled = isLoading;
        const btnText = this.submitBtn.querySelector('.btn-text');
        const btnLoading = this.submitBtn.querySelector('.btn-loading');
        
        if (isLoading) {
            btnText.style.display = 'none';
            btnLoading.style.display = 'inline';
        } else {
            btnText.style.display = 'inline';
            btnLoading.style.display = 'none';
        }
    }

    showError(message) {
        this.clearMessages();
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error-message';
        errorDiv.style.cssText = `
            background: #fee2e2;
            color: #991b1b;
            padding: 1rem;
            border-radius: 8px;
            margin-bottom: 1rem;
            border: 1px solid #fecaca;
        `;
        errorDiv.textContent = message;
        this.form.insertBefore(errorDiv, this.submitBtn);
    }

    showSuccessMessage(message) {
        this.clearMessages();
        const successDiv = document.createElement('div');
        successDiv.className = 'success-message';
        successDiv.style.cssText = `
            background: #d1fae5;
            color: #065f46;
            padding: 1rem;
            border-radius: 8px;
            margin-top: 1rem;
            border: 1px solid #a7f3d0;
        `;
        successDiv.textContent = message;
        this.resultsSection.appendChild(successDiv);
    }

    clearError() {
        this.clearMessages();
    }

    clearMessages() {
        const existingMessages = document.querySelectorAll('.error-message, .success-message');
        existingMessages.forEach(msg => msg.remove());
    }

    simulateDelay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    generateSessionId() {
        return 'session-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
    }
}

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    new ConsensusUI();
});
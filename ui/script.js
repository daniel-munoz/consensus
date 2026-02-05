class ConsensusUI {
    constructor() {
        this.apiBaseUrl = window.location.origin;
        this.pollingInterval = null;
        this.providers = [];
        this.initializeElements();
        this.attachEventListeners();
        this.loadProviders();
    }

    initializeElements() {
        this.form = document.getElementById('consensus-form');
        this.promptInput = document.getElementById('prompt');
        this.masterPromptToggle = document.getElementById('use-master-prompt');
        this.masterProviderSection = document.getElementById('master-provider-section');
        this.masterProviderSelect = document.getElementById('master-provider');
        this.responseProvidersContainer = document.getElementById('response-providers-container');
        this.emailInput = document.getElementById('email-recipients');
        this.submitBtn = document.getElementById('submit-btn');
        this.resultsSection = document.getElementById('results');
        this.resultsContainer = document.getElementById('results-container');
    }

    attachEventListeners() {
        this.form.addEventListener('submit', (e) => this.handleSubmit(e));
        this.masterPromptToggle.addEventListener('change', () => this.toggleMasterPrompt());
    }

    async loadProviders() {
        try {
            this.submitBtn.disabled = true;
            this.submitBtn.querySelector('.btn-text').textContent = 'Loading providers...';

            const response = await fetch(`${this.apiBaseUrl}/api/providers`);
            if (!response.ok) {
                throw new Error(`Failed to load providers: HTTP ${response.status}`);
            }

            const data = await response.json();
            this.providers = data.providers || [];

            if (this.providers.length === 0) {
                this.showError('No providers configured. Please check your configuration.');
                return;
            }

            this.populateProviders();
            this.submitBtn.disabled = false;
            this.submitBtn.querySelector('.btn-text').textContent = 'Send Request';
        } catch (error) {
            console.error('Failed to load providers:', error);
            this.showError(`Failed to load providers: ${error.message}`);
        }
    }

    populateProviders() {
        // Populate master provider dropdown
        this.masterProviderSelect.innerHTML = '';
        this.providers.forEach((provider, index) => {
            const option = document.createElement('option');
            option.value = provider;
            option.textContent = this.formatProviderName(provider);
            if (index === 0) option.selected = true;
            this.masterProviderSelect.appendChild(option);
        });

        // Populate response provider checkboxes
        this.responseProvidersContainer.innerHTML = '';
        this.providers.forEach(provider => {
            const label = document.createElement('label');
            const checkbox = document.createElement('input');
            checkbox.type = 'checkbox';
            checkbox.value = provider;
            checkbox.checked = true;
            checkbox.addEventListener('change', () => this.validateProviders());

            label.appendChild(checkbox);
            label.appendChild(document.createTextNode(` ${this.formatProviderName(provider)}`));
            this.responseProvidersContainer.appendChild(label);
        });
    }

    formatProviderName(provider) {
        // Capitalize first letter of each word
        return provider
            .split(/[-_]/)
            .map(word => word.charAt(0).toUpperCase() + word.slice(1))
            .join(' ');
    }

    getResponseProviderCheckboxes() {
        return this.responseProvidersContainer.querySelectorAll('input[type="checkbox"]');
    }

    toggleMasterPrompt() {
        const isEnabled = this.masterPromptToggle.checked;
        this.masterProviderSection.style.display = isEnabled ? 'block' : 'none';
    }

    validateProviders() {
        const checkboxes = this.getResponseProviderCheckboxes();
        const selectedProviders = Array.from(checkboxes).filter(cb => cb.checked);

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
            this.setLoadingState(false);
        }
    }

    collectFormData() {
        const checkboxes = this.getResponseProviderCheckboxes();
        const selectedProviders = Array.from(checkboxes)
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
        // Clear any existing polling interval to prevent race conditions
        if (this.pollingInterval) {
            clearInterval(this.pollingInterval);
            this.pollingInterval = null;
        }

        this.clearResults();

        // Show initial pending states
        if (data.useMasterPrompt) {
            this.addResult({
                provider: data.masterProvider,
                type: 'master-prompt',
                status: 'pending',
                content: 'Optimizing your prompt...'
            });
        }

        data.responseProviders.forEach(provider => {
            this.addResult({
                provider: provider,
                type: 'response',
                status: 'pending',
                content: `Waiting for ${this.formatProviderName(provider)} response...`
            });
        });

        // Submit request to API
        const response = await fetch(`${this.apiBaseUrl}/api/consensus`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || `HTTP ${response.status}`);
        }

        const { sessionId } = await response.json();

        // Start polling for results
        this.pollSession(sessionId, data);
    }

    async pollSession(sessionId, originalData) {
        const poll = async () => {
            try {
                const response = await fetch(`${this.apiBaseUrl}/api/session/${sessionId}`);

                if (!response.ok) {
                    const errorData = await response.json().catch(() => ({}));
                    throw new Error(errorData.error || `HTTP ${response.status}`);
                }

                const session = await response.json();
                this.updateFromSession(session, originalData);

                if (session.status === 'complete' || session.status === 'error') {
                    this.setLoadingState(false);
                    if (this.pollingInterval) {
                        clearInterval(this.pollingInterval);
                        this.pollingInterval = null;
                    }

                    // Show email notification if applicable
                    if (session.status === 'complete' && originalData.emailRecipients.length > 0) {
                        this.showSuccessMessage(`Email notifications sent to: ${originalData.emailRecipients.join(', ')}`);
                    }

                    if (session.status === 'error' && session.error) {
                        this.showError(`Session error: ${session.error}`);
                    }
                }
            } catch (error) {
                console.error('Polling error:', error);
                this.showError(`Failed to get session status: ${error.message}`);
                this.setLoadingState(false);
                if (this.pollingInterval) {
                    clearInterval(this.pollingInterval);
                    this.pollingInterval = null;
                }
            }
        };

        // Poll immediately, then every 2 seconds
        await poll();
        this.pollingInterval = setInterval(poll, 2000);
    }

    updateFromSession(session, originalData) {
        // Update master prompt if present
        if (session.masterPrompt) {
            const masterProvider = originalData.masterProvider;
            this.updateResult(masterProvider, 'master-prompt', {
                status: session.masterPrompt.status,
                content: session.masterPrompt.status === 'complete'
                    ? `Optimized prompt created using ${this.formatProviderName(masterProvider)}:\n\n${session.masterPrompt.content}`
                    : session.masterPrompt.status === 'error'
                        ? `Error: ${session.masterPrompt.error}`
                        : 'Optimizing your prompt...'
            });
        }

        // Update provider responses
        for (const [providerName, providerResponse] of Object.entries(session.responses)) {
            let content;
            if (providerResponse.status === 'complete') {
                content = providerResponse.content;
            } else if (providerResponse.status === 'error') {
                content = `Error: ${providerResponse.error}`;
            } else {
                content = `Waiting for ${this.formatProviderName(providerName)} response...`;
            }

            this.updateResult(providerName, 'response', {
                status: providerResponse.status,
                content: content
            });
        }
    }

    // Validate status value to prevent CSS class injection
    isValidStatus(status) {
        return ['pending', 'complete', 'error'].includes(status);
    }

    addResult(result) {
        const resultElement = document.createElement('div');
        resultElement.className = 'result-card';
        resultElement.dataset.provider = result.provider;
        resultElement.dataset.type = result.type;

        // Create header container
        const headerElement = document.createElement('div');
        headerElement.className = 'result-header';

        // Provider name
        const providerNameElement = document.createElement('span');
        providerNameElement.className = 'provider-name';
        const providerLabel = this.formatProviderName(result.provider) +
            (result.type === 'master-prompt' ? ' (Master Prompt)' : '');
        providerNameElement.textContent = providerLabel;

        // Status indicator
        const statusElement = document.createElement('span');
        const statusText = result.status.charAt(0).toUpperCase() + result.status.slice(1);
        statusElement.textContent = statusText;
        statusElement.className = 'status-indicator';
        if (this.isValidStatus(result.status)) {
            statusElement.classList.add(`status-${result.status}`);
        }

        headerElement.appendChild(providerNameElement);
        headerElement.appendChild(statusElement);

        // Content container
        const contentElement = document.createElement('div');
        contentElement.className = 'result-content';
        contentElement.textContent = result.content;

        resultElement.appendChild(headerElement);
        resultElement.appendChild(contentElement);

        this.resultsContainer.appendChild(resultElement);
    }

    updateResult(provider, type, updates) {
        // Use safer DOM query by iterating over children instead of querySelector with unescaped values
        const resultElement = Array.from(this.resultsContainer.children).find(
            (el) => el.dataset.provider === provider && el.dataset.type === type
        );
        if (!resultElement) return;

        if (updates.status) {
            const statusElement = resultElement.querySelector('.status-indicator');
            statusElement.textContent = updates.status.charAt(0).toUpperCase() + updates.status.slice(1);
            statusElement.className = 'status-indicator';
            if (this.isValidStatus(updates.status)) {
                statusElement.classList.add(`status-${updates.status}`);
            }
        }

        if (updates.content) {
            const contentElement = resultElement.querySelector('.result-content');
            contentElement.textContent = updates.content;
        }
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
}

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    new ConsensusUI();
});

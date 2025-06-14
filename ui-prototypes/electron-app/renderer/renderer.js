const { ipcRenderer } = require('electron');

class ConsensusDesktopApp {
    constructor() {
        this.currentResults = [];
        this.isProcessing = false;
        
        this.initializeElements();
        this.attachEventListeners();
        this.loadSettings();
    }

    initializeElements() {
        // Input elements
        this.promptInput = document.getElementById('prompt-input');
        this.masterPromptToggle = document.getElementById('master-prompt-toggle');
        this.masterPromptConfig = document.getElementById('master-prompt-config');
        this.masterProviderSelect = document.getElementById('master-provider');
        this.providerCheckboxes = document.querySelectorAll('.provider-card input[type="checkbox"]');
        this.emailInput = document.getElementById('email-input');
        
        // Button elements
        this.sendBtn = document.getElementById('send-btn');
        this.clearBtn = document.getElementById('clear-btn');
        this.settingsBtn = document.getElementById('settings-btn');
        this.saveResultsBtn = document.getElementById('save-results-btn');
        this.exportResultsBtn = document.getElementById('export-results-btn');
        
        // Panel elements
        this.resultsPanel = document.getElementById('results-panel');
        this.resultsContent = document.getElementById('results-content');
        
        // Modal elements
        this.settingsModal = document.getElementById('settings-modal');
        this.settingsClose = document.getElementById('settings-close');
        this.settingsCancel = document.getElementById('settings-cancel');
        this.settingsSave = document.getElementById('settings-save');
        
        // Settings form elements
        this.openaiKeyInput = document.getElementById('openai-key');
        this.anthropicKeyInput = document.getElementById('anthropic-key');
        this.geminiKeyInput = document.getElementById('gemini-key');
        this.smtpHostInput = document.getElementById('smtp-host');
        this.smtpPortInput = document.getElementById('smtp-port');
        this.fromEmailInput = document.getElementById('from-email');
    }

    attachEventListeners() {
        // Form submission
        this.sendBtn.addEventListener('click', (e) => this.handleSubmit(e));
        this.clearBtn.addEventListener('click', () => this.clearForm());
        
        // Master prompt toggle
        this.masterPromptToggle.addEventListener('change', () => this.toggleMasterPrompt());
        
        // Provider validation
        this.providerCheckboxes.forEach(checkbox => {
            checkbox.addEventListener('change', () => this.validateProviders());
        });
        
        // Settings modal
        this.settingsBtn.addEventListener('click', () => this.showSettings());
        this.settingsClose.addEventListener('click', () => this.hideSettings());
        this.settingsCancel.addEventListener('click', () => this.hideSettings());
        this.settingsSave.addEventListener('click', () => this.saveSettings());
        
        // Results actions
        this.saveResultsBtn.addEventListener('click', () => this.saveResults());
        this.exportResultsBtn.addEventListener('click', () => this.exportResults());
        
        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => this.handleKeyboard(e));
        
        // IPC Menu events
        ipcRenderer.on('menu-new-request', () => this.clearForm());
        ipcRenderer.on('menu-send-request', () => this.handleSubmit());
        ipcRenderer.on('menu-clear-results', () => this.clearResults());
        ipcRenderer.on('menu-settings', () => this.showSettings());
        ipcRenderer.on('menu-save-results', (event, filePath) => this.saveResultsToFile(filePath));
        
        // Modal click outside to close
        this.settingsModal.addEventListener('click', (e) => {
            if (e.target === this.settingsModal) {
                this.hideSettings();
            }
        });
    }

    toggleMasterPrompt() {
        const isEnabled = this.masterPromptToggle.checked;
        this.masterPromptConfig.style.display = isEnabled ? 'block' : 'none';
    }

    validateProviders() {
        const selectedProviders = Array.from(this.providerCheckboxes)
            .filter(cb => cb.checked);
        
        const isValid = selectedProviders.length > 0 && this.promptInput.value.trim();
        this.sendBtn.disabled = !isValid || this.isProcessing;
    }

    async handleSubmit(e) {
        if (e) e.preventDefault();
        
        const formData = this.collectFormData();
        if (!this.validateForm(formData)) return;

        this.setLoadingState(true);
        this.showResults();
        
        try {
            await this.processRequest(formData);
        } catch (error) {
            await this.showError('Request failed', error.message);
        } finally {
            this.setLoadingState(false);
        }
    }

    collectFormData() {
        const selectedProviders = Array.from(this.providerCheckboxes)
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
            this.showError('Validation Error', 'Please enter a prompt');
            return false;
        }

        if (data.responseProviders.length === 0) {
            this.showError('Validation Error', 'Please select at least one response provider');
            return false;
        }

        // Validate email format if provided
        if (data.emailRecipients.length > 0) {
            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            const invalidEmails = data.emailRecipients.filter(email => !emailRegex.test(email));
            if (invalidEmails.length > 0) {
                this.showError('Validation Error', `Invalid email addresses: ${invalidEmails.join(', ')}`);
                return false;
            }
        }

        return true;
    }

    async processRequest(data) {
        const sessionId = this.generateSessionId();
        this.clearResultsContent();
        this.currentResults = [];
        
        // Process master prompt if enabled
        if (data.useMasterPrompt) {
            const masterId = this.addResult({
                id: `master-${sessionId}`,
                provider: data.masterProvider,
                type: 'master-prompt',
                status: 'loading',
                content: 'Optimizing your prompt...'
            });

            await this.simulateDelay(1500);
            
            this.updateResult(masterId, {
                status: 'complete',
                content: this.generateMasterPromptResult(data.masterProvider, data.prompt)
            });
        }

        // Process provider responses
        const promises = data.responseProviders.map(async (providerId) => {
            const resultId = this.addResult({
                id: `${providerId}-${sessionId}`,
                provider: providerId,
                type: 'response',
                status: 'loading',
                content: `Waiting for ${this.getProviderName(providerId)} response...`
            });

            // Simulate varying response times
            const delay = Math.random() * 3000 + 1000;
            await this.simulateDelay(delay);

            const mockResponse = this.generateMockResponse(providerId, data.prompt);
            
            this.updateResult(resultId, {
                status: 'complete',
                content: mockResponse
            });
        });

        await Promise.all(promises);

        // Simulate email notification
        if (data.emailRecipients.length > 0) {
            await this.simulateDelay(500);
            await this.showInfo('Email Sent', `Notifications sent to: ${data.emailRecipients.join(', ')}`);
        }
    }

    addResult(result) {
        this.currentResults.push(result);
        this.renderResult(result);
        return result.id;
    }

    updateResult(resultId, updates) {
        const resultIndex = this.currentResults.findIndex(r => r.id === resultId);
        if (resultIndex === -1) return;

        this.currentResults[resultIndex] = { ...this.currentResults[resultIndex], ...updates };
        this.renderResult(this.currentResults[resultIndex]);
    }

    renderResult(result) {
        const existingCard = document.querySelector(`[data-result-id="${result.id}"]`);
        
        const cardHTML = `
            <div class="result-card ${result.type === 'master-prompt' ? 'master-prompt' : ''}" data-result-id="${result.id}">
                <div class="result-header">
                    <div class="result-provider">
                        <div class="provider-icon ${result.provider}"></div>
                        <span>${this.getProviderName(result.provider)} ${result.type === 'master-prompt' ? '(Master Prompt)' : ''}</span>
                    </div>
                    <div class="result-status">
                        <span class="status-badge ${result.status}">${this.capitalizeFirst(result.status)}</span>
                    </div>
                </div>
                <div class="result-content">${result.content}</div>
            </div>
        `;

        if (existingCard) {
            existingCard.outerHTML = cardHTML;
        } else {
            this.resultsContent.insertAdjacentHTML('beforeend', cardHTML);
        }
    }

    generateMasterPromptResult(provider, originalPrompt) {
        return `✨ **Optimized Prompt Created using ${this.getProviderName(provider)}**

**Original Request:**
"${originalPrompt}"

**Enhanced Version:**
"Provide a comprehensive, well-structured analysis of: ${originalPrompt}. Include multiple perspectives, practical applications, and actionable insights. Consider both immediate and long-term implications, potential challenges, and success metrics. Structure your response with clear headings and bullet points for maximum clarity."

**Optimization Applied:**
- Enhanced clarity and specificity
- Added structure requirements
- Included multiple perspective request
- Specified actionable output format`;
    }

    generateMockResponse(provider, prompt) {
        const responses = {
            openai: `**OpenAI GPT-4 Response:**

I'll provide a comprehensive analysis of your request: "${prompt}"

**Strategic Overview:**
• Context analysis and requirement identification
• Multi-faceted approach consideration
• Risk assessment and mitigation planning

**Detailed Analysis:**
1. **Current State Assessment**
   - Key factors and constraints
   - Available resources and capabilities
   - Market/environmental conditions

2. **Solution Framework**
   - Primary approach recommendations
   - Alternative strategies and contingencies
   - Implementation timeline and milestones

3. **Expected Outcomes**
   - Short-term deliverables and metrics
   - Long-term strategic benefits
   - Success indicators and KPIs

**Next Steps:**
- Immediate action items with ownership
- Resource requirements and allocation
- Monitoring and adjustment protocols

This analysis leverages GPT-4's comprehensive reasoning capabilities to provide structured, actionable insights.`,

            anthropic: `**Claude's Thoughtful Analysis:**

Thank you for this interesting question about "${prompt}". I'll approach this systematically and ethically.

**Understanding Your Request:**
I want to ensure I'm addressing the core of what you're asking while considering various perspectives and implications.

**Comprehensive Breakdown:**

🔍 **Analysis Framework:**
- Stakeholder impact assessment
- Ethical considerations and constraints
- Practical implementation challenges
- Long-term sustainability factors

📊 **Key Insights:**
- Evidence-based recommendations
- Balanced perspective on trade-offs
- Risk mitigation strategies
- Success probability assessment

🎯 **Actionable Recommendations:**

**Phase 1: Foundation (Immediate)**
- Validate core assumptions
- Gather stakeholder input
- Establish success metrics

**Phase 2: Implementation (Short-term)**
- Execute pilot programs
- Monitor and adjust approach
- Scale successful elements

**Phase 3: Optimization (Long-term)**
- Continuous improvement cycles
- Expand scope and impact
- Knowledge transfer and documentation

I strive to provide helpful, honest, and harmless guidance that considers multiple viewpoints and potential consequences.`,

            gemini: `**Gemini Multi-Modal Analysis:**

🌟 **Comprehensive Response to: "${prompt}"**

**Executive Summary:**
Leveraging Google's advanced AI capabilities to provide data-driven insights and innovative solutions.

📈 **Data-Driven Insights:**
- **Pattern Recognition:** Historical trends and analogous cases
- **Predictive Analytics:** Likely scenarios and outcome probabilities  
- **Optimization Opportunities:** Efficiency improvements and innovations

🔬 **Technical Deep-Dive:**

**Architecture & Design:**
- Scalable solution framework
- Integration considerations
- Performance optimization strategies

**Innovation Opportunities:**
- Emerging technology applications
- Creative problem-solving approaches
- Future-proofing strategies

🚀 **Implementation Roadmap:**

**Sprint 1 (Weeks 1-2): Discovery & Planning**
- Stakeholder alignment workshops
- Technical feasibility assessment
- Resource requirement analysis

**Sprint 2 (Weeks 3-4): Prototype Development**
- MVP creation and testing
- User feedback integration
- Performance baseline establishment

**Sprint 3 (Weeks 5-6): Scale & Optimize**
- Production deployment
- Monitoring and analytics setup
- Continuous improvement framework

**Competitive Advantages:**
- Leverages cutting-edge AI/ML capabilities
- Data-driven decision making
- Scalable and adaptable architecture
- Real-time optimization and learning

This response demonstrates Gemini's strength in combining analytical rigor with innovative thinking and practical implementation guidance.`
        };

        return responses[provider] || `Mock response from ${this.getProviderName(provider)} for: "${prompt}"`;
    }

    getProviderName(providerId) {
        const names = {
            openai: 'OpenAI',
            anthropic: 'Anthropic',
            gemini: 'Google Gemini'
        };
        return names[providerId] || providerId.toUpperCase();
    }

    capitalizeFirst(str) {
        return str.charAt(0).toUpperCase() + str.slice(1);
    }

    clearForm() {
        this.promptInput.value = '';
        this.emailInput.value = '';
        this.clearResults();
        this.promptInput.focus();
    }

    clearResults() {
        this.resultsPanel.style.display = 'none';
        this.clearResultsContent();
    }

    clearResultsContent() {
        this.resultsContent.innerHTML = '';
        this.currentResults = [];
    }

    showResults() {
        this.resultsPanel.style.display = 'flex';
    }

    setLoadingState(isLoading) {
        this.isProcessing = isLoading;
        this.sendBtn.disabled = isLoading;
        
        const btnText = this.sendBtn.querySelector('.btn-text');
        const btnLoading = this.sendBtn.querySelector('.btn-loading');
        
        if (isLoading) {
            btnText.style.display = 'none';
            btnLoading.style.display = 'flex';
        } else {
            btnText.style.display = 'inline';
            btnLoading.style.display = 'none';
        }
    }

    // Settings Management
    showSettings() {
        this.settingsModal.style.display = 'flex';
        this.openaiKeyInput.focus();
    }

    hideSettings() {
        this.settingsModal.style.display = 'none';
    }

    loadSettings() {
        // Load from localStorage or config file
        const settings = JSON.parse(localStorage.getItem('consensus-settings') || '{}');
        
        this.openaiKeyInput.value = settings.openaiKey || '';
        this.anthropicKeyInput.value = settings.anthropicKey || '';
        this.geminiKeyInput.value = settings.geminiKey || '';
        this.smtpHostInput.value = settings.smtpHost || 'smtp.gmail.com';
        this.smtpPortInput.value = settings.smtpPort || '587';
        this.fromEmailInput.value = settings.fromEmail || '';
    }

    saveSettings() {
        const settings = {
            openaiKey: this.openaiKeyInput.value,
            anthropicKey: this.anthropicKeyInput.value,
            geminiKey: this.geminiKeyInput.value,
            smtpHost: this.smtpHostInput.value,
            smtpPort: this.smtpPortInput.value,
            fromEmail: this.fromEmailInput.value
        };
        
        localStorage.setItem('consensus-settings', JSON.stringify(settings));
        this.hideSettings();
        this.showInfo('Settings Saved', 'Your configuration has been saved successfully.');
    }

    // File Operations
    async saveResults() {
        if (this.currentResults.length === 0) {
            await this.showInfo('No Results', 'No results to save. Please run a request first.');
            return;
        }

        const data = {
            timestamp: new Date().toISOString(),
            results: this.currentResults
        };

        const result = await ipcRenderer.invoke('save-file', 
            `consensus-results-${Date.now()}.json`, 
            JSON.stringify(data, null, 2)
        );

        if (result.success) {
            await this.showInfo('Results Saved', 'Results have been saved successfully.');
        } else {
            await this.showError('Save Failed', result.error);
        }
    }

    async exportResults() {
        if (this.currentResults.length === 0) {
            await this.showInfo('No Results', 'No results to export. Please run a request first.');
            return;
        }

        // Create a readable text format
        let exportText = `Consensus AI Results - ${new Date().toLocaleString()}\n`;
        exportText += '='.repeat(50) + '\n\n';

        this.currentResults.forEach((result, index) => {
            exportText += `${index + 1}. ${this.getProviderName(result.provider)}`;
            if (result.type === 'master-prompt') {
                exportText += ' (Master Prompt)';
            }
            exportText += `\n${'-'.repeat(30)}\n`;
            exportText += `${result.content}\n\n`;
        });

        const result = await ipcRenderer.invoke('save-file', 
            `consensus-export-${Date.now()}.txt`, 
            exportText
        );

        if (result.success) {
            await this.showInfo('Export Complete', 'Results have been exported to text file.');
        } else {
            await this.showError('Export Failed', result.error);
        }
    }

    // Keyboard Shortcuts
    handleKeyboard(e) {
        if (e.ctrlKey || e.metaKey) {
            switch (e.key) {
                case 'Enter':
                    e.preventDefault();
                    this.handleSubmit();
                    break;
                case 'n':
                    e.preventDefault();
                    this.clearForm();
                    break;
                case 'r':
                    e.preventDefault();
                    this.clearResults();
                    break;
                case 's':
                    e.preventDefault();
                    this.saveResults();
                    break;
                case ',':
                    e.preventDefault();
                    this.showSettings();
                    break;
            }
        }
        
        if (e.key === 'Escape') {
            this.hideSettings();
        }
    }

    // Utility Functions
    generateSessionId() {
        return 'session_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }

    simulateDelay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    async showError(title, message) {
        return ipcRenderer.invoke('show-error-dialog', title, message);
    }

    async showInfo(title, message) {
        return ipcRenderer.invoke('show-info-dialog', title, message);
    }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new ConsensusDesktopApp();
});
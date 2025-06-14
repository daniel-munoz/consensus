import React, { useState, useCallback } from 'react';
import { Switch } from '@headlessui/react';
import { 
  PaperAirplaneIcon, 
  CogIcon, 
  CheckCircleIcon, 
  ClockIcon,
  ExclamationCircleIcon,
  SparklesIcon
} from '@heroicons/react/24/outline';
import './App.css';

const PROVIDERS = [
  { id: 'openai', name: 'OpenAI', color: 'bg-green-500' },
  { id: 'anthropic', name: 'Anthropic', color: 'bg-orange-500' },
  { id: 'gemini', name: 'Gemini', color: 'bg-blue-500' }
];

function App() {
  const [prompt, setPrompt] = useState('');
  const [useMasterPrompt, setUseMasterPrompt] = useState(true);
  const [masterProvider, setMasterProvider] = useState('openai');
  const [selectedProviders, setSelectedProviders] = useState(['openai', 'anthropic', 'gemini']);
  const [emailRecipients, setEmailRecipients] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [results, setResults] = useState({});
  const [showResults, setShowResults] = useState(false);

  const handleProviderToggle = useCallback((providerId) => {
    setSelectedProviders(prev => 
      prev.includes(providerId) 
        ? prev.filter(id => id !== providerId)
        : [...prev, providerId]
    );
  }, []);

  const generateMockResponse = useCallback((provider, userPrompt) => {
    const responses = {
      openai: `**GPT-4 Analysis of: "${userPrompt}"**

I'll approach this systematically:

**Key Considerations:**
• Context analysis and requirement gathering
• Multiple solution pathways evaluation  
• Risk assessment and mitigation strategies
• Implementation feasibility

**Recommended Approach:**
1. **Immediate Actions:** Start with foundational research and stakeholder alignment
2. **Medium-term Strategy:** Develop iterative implementation plan with feedback loops
3. **Long-term Vision:** Scale and optimize based on measured outcomes

**Success Metrics:**
- Quantifiable performance indicators
- User satisfaction and engagement
- Resource efficiency and ROI

This response leverages OpenAI's structured analytical framework for comprehensive problem-solving.`,

      anthropic: `**Claude's Thoughtful Response to: "${userPrompt}"**

Thank you for this interesting question. Let me break this down thoughtfully:

**Understanding the Request:**
I want to ensure I'm addressing exactly what you're looking for, considering both explicit and implicit aspects of your query.

**Analysis Framework:**
- **Ethical Considerations:** What are the moral implications?
- **Stakeholder Impact:** Who is affected and how?
- **Practical Constraints:** What limitations should we consider?

**Balanced Perspective:**
• **Strengths:** Clear benefits and positive outcomes
• **Challenges:** Potential obstacles and how to navigate them
• **Alternatives:** Other approaches worth considering

**Actionable Next Steps:**
1. Validate assumptions through research
2. Engage relevant stakeholders early
3. Implement with careful monitoring and adjustment

I aim to be helpful while being honest about complexities and limitations.`,

      gemini: `**Gemini's Multi-Modal Analysis: "${userPrompt}"**

🎯 **Strategic Overview**
Your question intersects several domains that benefit from comprehensive analysis.

📊 **Data-Driven Insights**
- **Pattern Recognition:** Similar challenges in analogous contexts
- **Trend Analysis:** Current market/industry trajectories  
- **Predictive Modeling:** Likely outcomes under different scenarios

🔬 **Technical Deep-Dive**
- **Architecture Considerations:** Scalable and maintainable solutions
- **Integration Points:** How this connects to existing systems
- **Performance Optimization:** Efficiency and resource utilization

🚀 **Innovation Opportunities**
- **Emerging Technologies:** Cutting-edge approaches to consider
- **Creative Solutions:** Non-traditional methodologies
- **Future-Proofing:** Adaptability for evolving requirements

**Execution Roadmap:**
Phase 1: Foundation & Validation (Weeks 1-2)
Phase 2: Development & Testing (Weeks 3-6)  
Phase 3: Deployment & Optimization (Weeks 7-8)

Google's approach emphasizes comprehensive analysis with innovative, data-driven solutions.`
    };
    return responses[provider] || `Mock response from ${provider.toUpperCase()}`;
  }, []);

  const handleSubmit = useCallback(async (e) => {
    e.preventDefault();
    if (!prompt.trim() || selectedProviders.length === 0) return;

    setIsLoading(true);
    setShowResults(true);
    setResults({});

    try {
      // Simulate master prompt optimization
      if (useMasterPrompt) {
        setResults(prev => ({
          ...prev,
          masterPrompt: {
            provider: masterProvider,
            status: 'loading',
            content: 'Optimizing your prompt...',
            type: 'master-prompt'
          }
        }));

        await new Promise(resolve => setTimeout(resolve, 1500));
        
        setResults(prev => ({
          ...prev,
          masterPrompt: {
            ...prev.masterPrompt,
            status: 'complete',
            content: `✨ **Optimized Prompt Created**\n\nOriginal: "${prompt}"\n\n**Enhanced Version:**\n"Provide a comprehensive, well-structured analysis of: ${prompt}. Include multiple perspectives, practical applications, and actionable insights. Consider both immediate and long-term implications."`
          }
        }));
      }

      // Simulate provider responses
      const responsePromises = selectedProviders.map(async (providerId) => {
        setResults(prev => ({
          ...prev,
          [providerId]: {
            provider: providerId,
            status: 'loading',
            content: `Generating response from ${PROVIDERS.find(p => p.id === providerId)?.name}...`,
            type: 'response'
          }
        }));

        // Simulate varying response times
        const delay = Math.random() * 3000 + 1000;
        await new Promise(resolve => setTimeout(resolve, delay));

        const response = generateMockResponse(providerId, prompt);
        
        setResults(prev => ({
          ...prev,
          [providerId]: {
            ...prev[providerId],
            status: 'complete',
            content: response
          }
        }));
      });

      await Promise.all(responsePromises);

      // Simulate email notification
      if (emailRecipients.trim()) {
        await new Promise(resolve => setTimeout(resolve, 500));
        // You could show a toast notification here
      }

    } catch (error) {
      console.error('Error:', error);
    } finally {
      setIsLoading(false);
    }
  }, [prompt, selectedProviders, useMasterPrompt, masterProvider, emailRecipients, generateMockResponse]);

  const getStatusIcon = (status) => {
    switch (status) {
      case 'loading':
        return <ClockIcon className="h-5 w-5 text-yellow-500 animate-spin" />;
      case 'complete':
        return <CheckCircleIcon className="h-5 w-5 text-green-500" />;
      case 'error':
        return <ExclamationCircleIcon className="h-5 w-5 text-red-500" />;
      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-50 via-white to-cyan-50">
      <div className="max-w-4xl mx-auto p-6">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="flex items-center justify-center mb-4">
            <SparklesIcon className="h-10 w-10 text-indigo-600 mr-3" />
            <h1 className="text-4xl font-bold bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent">
              Consensus AI
            </h1>
          </div>
          <p className="text-xl text-gray-600">
            Get responses from multiple AI providers simultaneously
          </p>
        </div>

        {/* Main Form */}
        <div className="bg-white rounded-2xl shadow-xl p-8 mb-8">
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Prompt Input */}
            <div>
              <label htmlFor="prompt" className="block text-sm font-semibold text-gray-700 mb-2">
                Your Request
              </label>
              <textarea
                id="prompt"
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                className="w-full h-32 p-4 border-2 border-gray-200 rounded-xl focus:border-indigo-500 focus:ring-2 focus:ring-indigo-200 transition-colors resize-none"
                placeholder="Enter your question or request here..."
                required
              />
            </div>

            {/* Configuration Panel */}
            <div className="bg-gray-50 rounded-xl p-6 space-y-6">
              {/* Master Prompt Section */}
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900">Master Prompt Optimization</h3>
                    <p className="text-sm text-gray-600">Use AI to optimize your prompt before sending to providers</p>
                  </div>
                  <Switch
                    checked={useMasterPrompt}
                    onChange={setUseMasterPrompt}
                    className={`${useMasterPrompt ? 'bg-indigo-600' : 'bg-gray-200'} relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2`}
                  >
                    <span className={`${useMasterPrompt ? 'translate-x-6' : 'translate-x-1'} inline-block h-4 w-4 transform rounded-full bg-white transition-transform`} />
                  </Switch>
                </div>

                {useMasterPrompt && (
                  <div className="ml-4">
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Master prompt provider:
                    </label>
                    <select
                      value={masterProvider}
                      onChange={(e) => setMasterProvider(e.target.value)}
                      className="block w-48 p-2 border border-gray-300 rounded-lg focus:ring-indigo-500 focus:border-indigo-500"
                    >
                      {PROVIDERS.map(provider => (
                        <option key={provider.id} value={provider.id}>
                          {provider.name}
                        </option>
                      ))}
                    </select>
                  </div>
                )}
              </div>

              {/* Response Providers */}
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-gray-900">Response Providers</h3>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  {PROVIDERS.map(provider => (
                    <div
                      key={provider.id}
                      className={`p-4 rounded-lg border-2 cursor-pointer transition-all ${
                        selectedProviders.includes(provider.id)
                          ? 'border-indigo-500 bg-indigo-50'
                          : 'border-gray-200 hover:border-gray-300'
                      }`}
                      onClick={() => handleProviderToggle(provider.id)}
                    >
                      <div className="flex items-center space-x-3">
                        <div className={`w-4 h-4 rounded-full ${provider.color}`} />
                        <span className="font-medium text-gray-900">{provider.name}</span>
                        {selectedProviders.includes(provider.id) && (
                          <CheckCircleIcon className="h-5 w-5 text-indigo-600 ml-auto" />
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              {/* Email Notifications */}
              <div className="space-y-2">
                <label htmlFor="email" className="block text-sm font-medium text-gray-700">
                  Email Notifications (Optional)
                </label>
                <input
                  type="email"
                  id="email"
                  value={emailRecipients}
                  onChange={(e) => setEmailRecipients(e.target.value)}
                  className="w-full p-3 border border-gray-300 rounded-lg focus:ring-indigo-500 focus:border-indigo-500"
                  placeholder="email1@example.com, email2@example.com"
                  multiple
                />
              </div>
            </div>

            {/* Submit Button */}
            <button
              type="submit"
              disabled={isLoading || !prompt.trim() || selectedProviders.length === 0}
              className="w-full bg-gradient-to-r from-indigo-600 to-purple-600 text-white font-semibold py-4 px-6 rounded-xl hover:from-indigo-700 hover:to-purple-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center justify-center space-x-2"
            >
              {isLoading ? (
                <>
                  <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white" />
                  <span>Processing...</span>
                </>
              ) : (
                <>
                  <PaperAirplaneIcon className="h-5 w-5" />
                  <span>Send Request</span>
                </>
              )}
            </button>
          </form>
        </div>

        {/* Results Section */}
        {showResults && (
          <div className="space-y-6">
            <h2 className="text-2xl font-bold text-gray-900 flex items-center">
              <CogIcon className="h-6 w-6 mr-2" />
              Results
            </h2>

            {/* Master Prompt Result */}
            {results.masterPrompt && (
              <div className="bg-white rounded-xl shadow-lg p-6 border-l-4 border-purple-500">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center space-x-2">
                    <SparklesIcon className="h-5 w-5 text-purple-600" />
                    <h3 className="text-lg font-semibold text-gray-900">
                      Master Prompt ({PROVIDERS.find(p => p.id === results.masterPrompt.provider)?.name})
                    </h3>
                  </div>
                  {getStatusIcon(results.masterPrompt.status)}
                </div>
                <div className="prose prose-sm max-w-none">
                  <pre className="whitespace-pre-wrap text-gray-700 font-sans">
                    {results.masterPrompt.content}
                  </pre>
                </div>
              </div>
            )}

            {/* Provider Responses */}
            <div className="grid gap-6">
              {selectedProviders.map(providerId => {
                const provider = PROVIDERS.find(p => p.id === providerId);
                const result = results[providerId];
                
                return (
                  <div key={providerId} className="bg-white rounded-xl shadow-lg p-6">
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center space-x-3">
                        <div className={`w-4 h-4 rounded-full ${provider.color}`} />
                        <h3 className="text-lg font-semibold text-gray-900">
                          {provider.name}
                        </h3>
                      </div>
                      {result && getStatusIcon(result.status)}
                    </div>
                    
                    {result && (
                      <div className="prose prose-sm max-w-none">
                        <pre className="whitespace-pre-wrap text-gray-700 font-sans leading-relaxed">
                          {result.content}
                        </pre>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default App;
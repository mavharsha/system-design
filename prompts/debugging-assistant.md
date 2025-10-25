# Debugging Notes

## Purpose
Create structured debugging notes to systematically identify and fix issues.

## Prompt

You are creating debugging notes to help solve the problem methodically. Please structure your debugging analysis in note form.

**Follow this debugging process:**

1. **Problem Understanding**
   - What is the expected behavior?
   - What is the actual behavior?
   - When did the problem start occurring?
   - Can the problem be consistently reproduced?
   - What are the exact steps to reproduce?

2. **Initial Assessment**
   - What error messages or logs are available?
   - What was changed recently (code, config, infrastructure)?
   - What environment is affected (dev, staging, prod)?
   - What is the impact and severity?

3. **Hypothesis Generation**
   - List possible root causes
   - Rank hypotheses by likelihood
   - Consider common causes first (network, permissions, config)
   - Don't ignore uncommon but possible causes

4. **Investigation Strategy**
   - For each hypothesis, suggest specific tests or checks
   - Recommend logging or debugging tools to use
   - Suggest data to collect
   - Identify diagnostic commands to run

5. **Code Analysis**
   - Review relevant code sections
   - Check for common bug patterns:
     - Null/undefined handling
     - Race conditions
     - Off-by-one errors
     - Resource leaks
     - Incorrect assumptions
     - Edge cases not handled

6. **System Analysis**
   - Check system resources (memory, CPU, disk, network)
   - Verify configurations and environment variables
   - Check dependencies and versions
   - Review recent deployments or changes
   - Examine logs for patterns

7. **Data Analysis**
   - Verify data integrity
   - Check for data corruption
   - Look for unusual data patterns
   - Validate input/output

8. **Root Cause Identification**
   - Narrow down to the specific cause
   - Verify the root cause with evidence
   - Explain why this causes the observed behavior

9. **Solution Recommendations**
   - Immediate fix to resolve the issue
   - Temporary workaround if immediate fix is complex
   - Long-term solution to prevent recurrence
   - Testing steps to verify the fix

10. **Prevention Measures**
    - How to prevent this issue in the future
    - Monitoring/alerting to catch early
    - Code improvements or safeguards
    - Documentation updates needed

**Debugging Best Practices:**
- Use binary search/divide-and-conquer approach
- Change one thing at a time
- Verify assumptions with evidence
- Use rubber duck debugging (explain step by step)
- Check the obvious things first
- Look for patterns in logs/errors
- Consider timing and concurrency issues

**Create notes in this format:**

1. **Problem Summary**
   - What's broken (1-2 sentences)
   - Impact and severity

2. **Quick Checks** (do these first)
   - 3-4 most likely causes to check
   - Quick diagnostic commands

3. **Investigation Plan**
   - Step 1: [what to check and why]
   - Step 2: [next thing to check]
   - Step 3: [and so on...]

4. **Most Likely Root Cause**
   - Based on symptoms, what's probably wrong
   - Evidence supporting this theory
   - How to verify

5. **Fix Plan**
   - Immediate fix (step-by-step)
   - How to test the fix
   - Prevent recurrence (1-2 actions)

**Notes Style:**
- Be methodical but concise
- Use numbered steps
- Focus on most probable causes first
- Include specific commands/actions
- Keep it practical

Create debugging notes that lead to a solution quickly without overwhelming detail.


# Mathematical Formulation of Trajectory Variance Metrics

## 1. Tool Invocation Efficiency Ratio ($\eta_{\text{tools}}$)

Let the full sequence of tool calls executed by an agent during a task be denoted by:

$$\mathcal{T} = (t_1, t_2, \dots, t_K)$$

where $t_k$ represents the $k$-th tool invocation with its target resource and parameters.

Let $\mathcal{T}_{\text{essential}} \subseteq \mathcal{T}$ represent the subset of tool calls that directly contributed to resolving the task (for example, target file inspection, file edits, and verification test runs), excluding redundant reads and circular command retries.

The Tool Invocation Efficiency Ratio is defined as:

$$\eta_{\text{tools}} = \frac{|\mathcal{T}_{\text{essential}}|}{|\mathcal{T}|}, \quad \eta_{\text{tools}} \in (0, 1]$$

A regression is flagged when the candidate efficiency drops significantly below the baseline:

$$\Delta \eta_{\text{tools}} = \eta_{\text{candidate}} - \eta_{\text{baseline}} < -0.15$$

---

## 2. Tool Call Bloat Factor ($B_{\text{tools}}$)

When comparing candidate model $C$ and baseline model $B$ on the identical task $i$:

$$B_{\text{tools}}(i) = \frac{|\mathcal{T}_C(i)|}{|\mathcal{T}_B(i)|}$$

For a benchmark suite containing $M$ tasks:

$$\bar{B}_{\text{tools}} = \frac{1}{M} \sum_{i=1}^{M} B_{\text{tools}}(i)$$

Values of $\bar{B}_{\text{tools}} > 1.20$ indicate that the candidate model uses on average 20% more tool calls than the baseline to solve identical problems.

---

## 3. AST Surgical Precision Score ($S_{\text{surgical}}$)

Traditional line diff metrics reward or penalize code edits purely by line count. When autonomous agents modify code, line churn often includes broad file rewrites, re-indentation, or formatting changes that obscure functional edits.

Let:
- $L_{\text{diff}}$ be the total count of modified lines (insertions + deletions) reported by Git diff.
- $N_{\text{ast}}$ be the count of modified Abstract Syntax Tree (AST) functional nodes (declarations, statements, control flow blocks, expressions), excluding whitespace and standalone comments.

The Surgical Precision Score is defined as:

$$S_{\text{surgical}} = \frac{N_{\text{ast}}}{L_{\text{diff}} + 1}$$

### Score Interpretation
- $S_{\text{surgical}} \ge 0.70$: High surgical precision. Changes are targeted and localized to functional AST nodes.
- $0.40 \le S_{\text{surgical}} < 0.70$: Moderate precision. Contains minor formatting drift or intermediate rewrites.
- $S_{\text{surgical}} < 0.40$: Low precision. Indicates mass reformatting, wholesale file replacement, or cosmetic churn.

---

## 4. Cosmetic Churn Ratio ($C_{\text{cosmetic}}$)

Cosmetic churn measures the fraction of modified lines that do not alter the semantic AST:

$$C_{\text{cosmetic}} = \max\left(0, 1 - \frac{N_{\text{ast}}}{L_{\text{diff}} + 1}\right)$$

A candidate model exhibiting $C_{\text{cosmetic}} > 0.50$ is flagged for formatting instability.

---

## 5. Trajectory Edit Distance ($D_{\text{traj}}$)

To measure structural divergence between two decision sequences, tool invocations are mapped to discrete tokens:

$$\text{Token}(t_k) = \text{tool\_name} + \text{hash}(\text{normalized\_target})$$

Let $\mathbf{a} = (\text{Token}(t_{A,1}), \dots, \text{Token}(t_{A,n}))$ and $\mathbf{b} = (\text{Token}(t_{B,1}), \dots, \text{Token}(t_{B,m}))$.

The Trajectory Edit Distance is computed using Levenshtein distance:

$$D_{\text{traj}}(\mathbf{a}, \mathbf{b}) = \text{Levenshtein}(\mathbf{a}, \mathbf{b})$$

The normalized distance over maximum length is:

$$\bar{D}_{\text{traj}}(\mathbf{a}, \mathbf{b}) = \frac{D_{\text{traj}}(\mathbf{a}, \mathbf{b})}{\max(|\mathbf{a}|, |\mathbf{b}|)}$$

### Divergence Turn ($T_{\text{div}}$)
The divergence turn is defined as the smallest index $k$ such that:

$$T_{\text{div}} = \min \{ k \mid \text{Token}(t_{A,k}) \neq \text{Token}(t_{B,k}) \}$$

This pinpoints the precise turn where the two models adopted different operational strategies.

---

## 6. Anti-Pattern Detection Formulations

### 6.1 Redundant File Reads
A tool call $t_j = \text{read}(f)$ is classified as redundant if there exists a prior call $t_i = \text{read}(f)$ with $i < j$, such that no intervening write or edit tool call modified file $f$ between step $i$ and step $j$.

### 6.2 Command Doom Loop
A shell execution $t_j = \text{bash}(cmd)$ is classified as part of a doom loop if:
1. $t_j$ returned a non-zero exit code.
2. Identical command $cmd$ was executed at least twice in preceding turns with identical non-zero exit codes.
3. No intermediate file modification occurred between executions.

### 6.3 Shotgun Surgery
Let $F_{\text{target}}$ be the set of files directly relevant to the task benchmark. If the set of files modified by the agent $F_{\text{modified}}$ satisfies:

$$\frac{|F_{\text{modified}} \setminus F_{\text{target}}|}{|F_{\text{modified}}|} > 0.30$$

the trajectory is flagged for shotgun surgery.

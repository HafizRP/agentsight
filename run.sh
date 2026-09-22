#!/bin/bash
set -e

AGENT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="${PROJECT_DIR:-$(pwd)}"
COMMON_MANDATES="$AGENT_DIR/mandates_common.md"
TASKS_DIR="$AGENT_DIR/tasks"

echo "🔍 [AGENTSIGHT RUNNER] Memulai pipeline modular..."
echo "📁 Agent Dir: $AGENT_DIR"
echo "📁 Project Dir: $PROJECT_DIR"

# Pindah ke project directory
cd "$PROJECT_DIR"

# Pastikan contracts directory ada
mkdir -p "$AGENT_DIR/contracts"

run_step() {
    local task_id="$1"
    local spec_file="$2"
    local branch="$3"
    local retry_count=0
    local max_retries=3

    echo ""
    echo "=========================================================="
    echo "▶️  SPAWNING SUB-AGENT: $task_id ($spec_file)"
    echo "   Branch: $branch"
    echo "   Working Dir: $PROJECT_DIR"
    echo "=========================================================="

    # Pindah / Buat branch git
    if ! git checkout "$branch" 2>/dev/null; then
        git checkout -b "$branch"
    fi

    while [ $retry_count -lt $max_retries ]; do
        echo "⏳ Percobaan $(($retry_count + 1))/$max_retries..."

        # Gabungkan mandates universal + task instruction
        cat "$COMMON_MANDATES" "$spec_file" > .hermes_task_payload.tmp

        # Eksekusi oneshot mode Hermes (-z) dengan context terisolasi
        if hermes -z "$(cat .hermes_task_payload.tmp)"; then
            rm -f .hermes_task_payload.tmp
            echo "✅ [$task_id] SUKSES. Context dibersihkan dari RAM."

            # Commit jika ada perubahan
            if [ -n "$(git status --porcelain)" ]; then
                git add -A
                git commit -m "feat($branch): $task_id completed by sub-agent"
            fi

            return 0
        else
            retry_count=$(($retry_count + 1))
            echo "⚠️ [$task_id] Gagal pada percobaan $retry_count."
            sleep 3
        fi
    done

    rm -f .hermes_task_payload.tmp
    echo "❌ [CIRCUIT BREAKER] $task_id gagal setelah $max_retries percobaan! Pipeline berhenti."
    echo "Failure recorded at $(date)" > "$AGENT_DIR/FAILURE_${task_id}.log"
    exit 1
}

# ==============================================================
# Pipeline Execution DAG
# ==============================================================

# Phase 1: Backend (foundation — sequential)
run_step "TASK_1" "$TASKS_DIR/01_backend.md" "feat/backend-api"

# Phase 2: Scraper + Frontend (parallel — independent after backend)
echo ""
echo "🔀 LAUNCHING PARALLEL PHASE: TASK_2 + TASK_3"
echo ""

# Run TASK_2 in background
(
    run_step "TASK_2" "$TASKS_DIR/02_worker.md" "feat/scraper-engine"
) &
PID_TASK2=$!

# Run TASK_3 in foreground
run_step "TASK_3" "$TASKS_DIR/03_frontend.md" "feat/frontend-ui"

# Wait for TASK_2 background job
echo "⏳ Menunggu TASK_2 (Scraper) selesai..."
wait $PID_TASK2
TASK2_EXIT=$?

if [ $TASK2_EXIT -ne 0 ]; then
    echo "❌ TASK_2 (Scraper) gagal dengan exit code $TASK2_EXIT"
    exit 1
fi

echo "✅ PARALLEL PHASE SELESAI: TASK_2 + TASK_3"

# Phase 3: Code Review (gate — sequential)
run_step "TASK_4" "$TASKS_DIR/04_reviewer.md" "main"

# Phase 4: Infrastructure (sequential)
run_step "TASK_5" "$TASKS_DIR/05_infra.md" "infra/docker-deploy"

# Phase 5: Deployment & Verification (final — sequential)
run_step "TASK_6" "$TASKS_DIR/06_verify.md" "main"

echo ""
echo "=========================================================="
echo "🎉 SELURUH PIPELINE SELESAI DENGAN STATUS 100% PASS!"
echo "📁 Project: $(pwd)"
echo "🌐 App: http://localhost:${APP_PORT:-8084}"
echo "=========================================================="

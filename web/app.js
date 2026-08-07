"use strict";

const LADDER = [
  "naked_single",
  "hidden_single",
  "locked_candidates_pointing",
  "locked_candidates_claiming",
  "naked_subset",
  "hidden_subset",
  "x_wing",
  "swordfish",
  "jellyfish",
  "xy_wing",
  "xyz_wing",
  "w_wing",
  "simple_colouring"
];

const NAMES = {
  naked_single: "Naked single",
  hidden_single: "Hidden single",
  locked_candidates_pointing: "Locked candidates (pointing)",
  locked_candidates_claiming: "Locked candidates (claiming)",
  naked_subset: "Naked subset",
  hidden_subset: "Hidden subset",
  x_wing: "X-wing",
  swordfish: "Swordfish",
  jellyfish: "Jellyfish",
  xy_wing: "XY-wing",
  xyz_wing: "XYZ-wing",
  w_wing: "W-wing",
  simple_colouring: "Simple colouring"
};

const BANDS = {
  naked_single: "Easy",
  hidden_single: "Easy",
  locked_candidates_pointing: "Medium",
  locked_candidates_claiming: "Medium",
  naked_subset: "Medium",
  hidden_subset: "Medium",
  x_wing: "Hard",
  swordfish: "Hard",
  jellyfish: "Hard",
  xy_wing: "Hard",
  xyz_wing: "Expert",
  w_wing: "Expert",
  simple_colouring: "Expert"
};

const EXPLAIN = {
  naked_single:
    "A cell has exactly one candidate digit left, so that digit must go there. The most basic forced placement.",
  hidden_single:
    "Within one row, column, or box, only a single cell can still take a certain digit — so the digit must go in that cell, even if the cell has other candidates.",
  locked_candidates_pointing:
    "All of a digit's candidates inside a box line up on one row or column. The digit must land in that box, so it can be removed from the rest of that row or column outside the box.",
  locked_candidates_claiming:
    "All of a digit's candidates in a row or column fall inside a single box. The digit must be placed in that intersection, so it can be removed from the rest of the box.",
  naked_subset:
    "A group of N cells in one unit shares exactly the same N candidate digits between them, locking those digits to those cells. The digits can be removed from every other cell in the unit.",
  hidden_subset:
    "N digits are confined to the same N cells of a unit, so those cells can hold nothing else. Every other candidate is removed from those cells.",
  x_wing:
    "One digit is restricted to the same two columns in each of two rows (or the same two rows in two columns), forming a rectangle. It must take opposite corners, so it is removed from the rest of those columns or rows.",
  swordfish:
    "The X-wing pattern extended to three lines: a digit confined to the same three columns across three rows (or vice versa) must occupy one cell per line, so it is removed from all other cells in those columns or rows.",
  jellyfish:
    "The four-line version of the fish family: a digit confined to the same four columns across four rows (or vice versa) is eliminated from every other cell of those four covering lines.",
  xy_wing:
    "Three bi-value cells form a pivot (XY) and two pincers (XZ and YZ). Whichever value the pivot takes, one pincer becomes Z — so Z is removed from any cell that sees both pincers.",
  xyz_wing:
    "Like an XY-wing, but the pivot holds all three candidates (XYZ). In every outcome the digit Z lands in the pivot or a pincer, so Z is removed from cells that see all three.",
  w_wing:
    "Two cells with the same candidate pair are joined by a strong link on one of the two digits. Either way, the other digit must occupy one of the pair, so it is removed from cells that see both of them.",
  simple_colouring:
    "Cells joined by strong links on one digit are alternately two-coloured. If one colour appears twice in a unit, or a cell sees both colours, that contradiction eliminates the losing colour's candidates."
};

const state = {
  events: [],
  baseGrid: null,
  step: 0,
  timer: 0
};

const byId = (id) => document.getElementById(id);

const gridEl = byId("grid");
const seedSelect = byId("seed-select");
const pasteInput = byId("paste-input");
const clearBtn = byId("clear-btn");
const solveBtn = byId("solve-btn");
const statusEl = byId("status");
const statsDifficulty = byId("stats-difficulty");
const histogram = byId("histogram");
const stepPos = byId("step-pos");
const stepDesc = byId("step-desc");
const eventLog = byId("event-log");
const hintEl = byId("hint");
const explainText = byId("explain-text");
const bandChip = byId("band-chip");
const btnFirst = byId("btn-first");
const btnPrev = byId("btn-prev");
const btnPlay = byId("btn-play");
const btnNext = byId("btn-next");
const btnLast = byId("btn-last");

const CELL_CLASSES = ["given", "filled", "placement", "witness", "elimination"];

const cells = [];
for (let i = 0; i < 81; i++) {
  const cell = document.createElement("input");
  cell.type = "text";
  cell.className = "cell";
  cell.setAttribute("inputmode", "numeric");
  cell.setAttribute("maxlength", "1");
  cell.setAttribute("aria-label", "Row " + (Math.floor(i / 9) + 1) + ", Column " + ((i % 9) + 1));
  cell.dataset.index = String(i);
  cell.addEventListener("input", onCellInput);
  gridEl.append(cell);
  cells.push(cell);
}

function cellAt(row, col) {
  return cells[row * 9 + col];
}

function rc(row, col) {
  return "R" + (row + 1) + "C" + (col + 1);
}

function isBlank(ch) {
  return ch === "0" || ch === ".";
}

function onCellInput(e) {
  const el = e.target;
  const digits = el.value.match(/[1-9]/g);
  el.value = digits ? digits[digits.length - 1] : "";
  resetSolveState();
  if (el.value) {
    const next = cells[Number(el.dataset.index) + 1];
    if (next) {
      next.focus();
      next.select();
    }
  }
}

function gridString() {
  return cells.map((c) => c.value || "0").join("");
}

function setGrid(str) {
  for (let i = 0; i < 81; i++) {
    cells[i].value = isBlank(str[i]) ? "" : str[i];
  }
}

function setStatus(text) {
  statusEl.textContent = text;
}

function setMetrics(iterations, eventCount, candidateChecks, solveTimeMs) {
  byId("m-iterations").textContent = iterations;
  byId("m-events").textContent = eventCount;
  byId("m-checks").textContent = candidateChecks;
  byId("m-time").textContent = solveTimeMs;
}

function formatMs(ms) {
  return ms.toFixed(2) + " ms";
}

// --- Seed dropdown ---

async function loadPuzzles() {
  try {
    const res = await fetch("/v1/puzzles");
    if (!res.ok) {
      throw new Error("HTTP " + res.status);
    }
    const data = await res.json();
    for (const section of data.sections || []) {
      const group = document.createElement("optgroup");
      group.label = section.name;
      (section.puzzles || []).forEach((puzzle, i) => {
        const opt = document.createElement("option");
        opt.value = puzzle;
        opt.textContent = section.name + " #" + (i + 1);
        group.append(opt);
      });
      seedSelect.append(group);
    }
  } catch (err) {
    setStatus("Could not load seed puzzles: " + err.message);
  }
}

seedSelect.addEventListener("change", () => {
  if (!seedSelect.value) {
    return;
  }
  setGrid(seedSelect.value);
  pasteInput.value = "";
  resetSolveState();
});

// --- Paste box ---

pasteInput.addEventListener("input", () => {
  const raw = pasteInput.value.replace(/\s+/g, "");
  if (raw.length !== 81 || /[^0-9.]/.test(raw)) {
    return;
  }
  setGrid(raw);
  seedSelect.value = "";
  resetSolveState();
});

// --- Clear ---

clearBtn.addEventListener("click", () => {
  for (const cell of cells) {
    cell.value = "";
  }
  pasteInput.value = "";
  seedSelect.value = "";
  resetSolveState();
});

// --- Solve ---

solveBtn.addEventListener("click", requestSolve);

async function requestSolve() {
  const puzzle = gridString();
  solveBtn.disabled = true;
  setStatus("Solving…");
  try {
    const res = await fetch("/v1/solve", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ puzzle })
    });
    const data = await res.json();
    if (res.status === 200 || (res.status === 400 && data.status === "invalid_input")) {
      renderSolve(puzzle, data);
    } else {
      resetSolveState();
      setStatus("Error: " + (data.error || "HTTP " + res.status) + (data.code ? " (" + data.code + ")" : ""));
    }
  } catch (err) {
    resetSolveState();
    setStatus("Request failed: " + err.message);
  } finally {
    solveBtn.disabled = false;
  }
}

function renderSolve(puzzle, data) {
  stopPlay();
  const viewable = data.status !== "invalid_input";
  state.events = viewable && Array.isArray(data.events) ? data.events : [];
  state.baseGrid = viewable
    ? (typeof data.input === "string" && data.input.length === 81 ? data.input : puzzle)
    : null;
  setStatus(statusLine(data));
  setMetrics(
    String(data.iterations),
    String(data.eventCount),
    String(data.candidateChecks),
    formatMs(data.solveTimeMs)
  );
  renderStats(data);
  renderLog();
  if (state.baseGrid) {
    renderStep(state.events.length);
  } else {
    resetViewerPanels();
  }
}

function statusLine(data) {
  switch (data.status) {
    case "solved":
      return "Status: solved · Grade: " + data.grade;
    case "invalid_input":
      return "Status: invalid_input — not a valid sudoku input.";
    case "stalled":
      return "Status: stalled — the technique ladder found no further logical move.";
    case "unsolvable":
      return "Status: unsolvable — a cell was left with no possible digit.";
    default:
      return "Status: " + data.status;
  }
}

// --- Statistics panel ---

function renderStats(data) {
  histogram.replaceChildren();
  if (!data) {
    statsDifficulty.textContent = "";
    return;
  }
  statsDifficulty.textContent =
    data.status === "solved"
      ? "Difficulty: " + data.grade
      : "Difficulty: — (" + data.status + ")";
  const events = Array.isArray(data.events) ? data.events : [];
  const counts = {};
  for (const ev of events) {
    counts[ev.technique] = (counts[ev.technique] || 0) + 1;
  }
  const fired = LADDER.filter((t) => counts[t]);
  if (fired.length === 0) {
    const p = document.createElement("p");
    p.className = "histogram-empty";
    p.textContent = "No technique events.";
    histogram.append(p);
    return;
  }
  const max = Math.max(...fired.map((t) => counts[t]));
  for (const tech of fired) {
    const row = document.createElement("div");
    row.className = "bar-row";
    const label = document.createElement("span");
    label.textContent = NAMES[tech];
    const track = document.createElement("div");
    track.className = "bar-track";
    const bar = document.createElement("div");
    bar.className = "bar";
    bar.style.width = ((counts[tech] / max) * 100).toFixed(1) + "%";
    track.append(bar);
    const count = document.createElement("span");
    count.className = "bar-count";
    count.textContent = String(counts[tech]);
    row.append(label, track, count);
    histogram.append(row);
  }
}

// --- Event log ---

function renderLog() {
  eventLog.replaceChildren();
  state.events.forEach((ev, idx) => {
    const li = document.createElement("li");
    li.dataset.step = String(idx + 1);
    li.textContent = "#" + (idx + 1) + " " + ev.technique + " — " + effectSummary(ev, 2);
    eventLog.append(li);
  });
}

eventLog.addEventListener("click", (e) => {
  const li = e.target.closest("li");
  if (!li || !li.dataset.step) {
    return;
  }
  stopPlay();
  renderStep(Number(li.dataset.step));
});

function effectSummary(ev, cap) {
  if (ev.placement) {
    return "places " + ev.placement.digit + " at " + rc(ev.placement.row, ev.placement.col);
  }
  const elims = ev.eliminations || [];
  const parts = elims.slice(0, cap).map((el) => el.digit + " from " + rc(el.row, el.col));
  let text = "eliminates " + parts.join(", ");
  if (elims.length > cap) {
    text += " +" + (elims.length - cap) + " more";
  }
  return text;
}

// --- Step viewer ---

function renderStep(i) {
  const n = state.events.length;
  state.step = Math.max(0, Math.min(i, n));
  stepPos.textContent = "Step " + state.step + " / " + n;
  paintGrid();
  applyHighlights();
  renderPanels();
  syncLog();
  updateTransport();
}

function paintGrid() {
  const grid = state.step === 0 ? state.baseGrid : state.events[state.step - 1].gridAfter;
  for (let k = 0; k < 81; k++) {
    cells[k].value = isBlank(grid[k]) ? "" : grid[k];
    cells[k].classList.remove(...CELL_CLASSES);
    if (!isBlank(grid[k])) {
      cells[k].classList.add(isBlank(state.baseGrid[k]) ? "filled" : "given");
    }
  }
}

function applyHighlights() {
  if (state.step === 0) {
    return;
  }
  const ev = state.events[state.step - 1];
  for (const w of ev.witnessCells || []) {
    cellAt(w.row, w.col).classList.add("witness");
  }
  for (const el of ev.eliminations || []) {
    cellAt(el.row, el.col).classList.add("elimination");
  }
  if (ev.placement) {
    const target = cellAt(ev.placement.row, ev.placement.col);
    target.classList.remove("witness", "elimination");
    target.classList.add("placement");
  }
}

function renderPanels() {
  if (state.step === 0) {
    stepDesc.textContent =
      state.events.length > 0 ? "Input grid — press Play or Next to walk the solve." : "";
    hintEl.classList.remove("hidden");
    explainText.classList.add("hidden");
    bandChip.classList.add("hidden");
    return;
  }
  const ev = state.events[state.step - 1];
  stepDesc.textContent = (NAMES[ev.technique] || ev.technique) + " " + effectSummary(ev, 4);
  hintEl.classList.add("hidden");
  explainText.textContent = EXPLAIN[ev.technique] || "";
  explainText.classList.remove("hidden");
  bandChip.textContent = BANDS[ev.technique] || "";
  bandChip.classList.remove("hidden");
}

function syncLog() {
  const rows = eventLog.children;
  for (let k = 0; k < rows.length; k++) {
    rows[k].classList.toggle("active", k === state.step - 1);
  }
  if (state.step > 0 && rows[state.step - 1]) {
    rows[state.step - 1].scrollIntoView({ block: "nearest" });
  }
}

function updateTransport() {
  const n = state.events.length;
  btnFirst.disabled = state.step === 0;
  btnPrev.disabled = state.step === 0;
  btnPlay.disabled = n === 0;
  btnNext.disabled = state.step >= n;
  btnLast.disabled = state.step >= n;
}

btnFirst.addEventListener("click", () => {
  stopPlay();
  renderStep(0);
});
btnPrev.addEventListener("click", () => {
  stopPlay();
  renderStep(state.step - 1);
});
btnNext.addEventListener("click", () => {
  stopPlay();
  renderStep(state.step + 1);
});
btnLast.addEventListener("click", () => {
  stopPlay();
  renderStep(state.events.length);
});
btnPlay.addEventListener("click", () => {
  if (state.timer) {
    stopPlay();
    return;
  }
  if (state.events.length === 0) {
    return;
  }
  if (state.step >= state.events.length) {
    renderStep(0);
  }
  btnPlay.textContent = "Pause";
  state.timer = setInterval(() => {
    renderStep(state.step + 1);
    if (state.step >= state.events.length) {
      stopPlay();
    }
  }, 800);
});

function stopPlay() {
  if (state.timer) {
    clearInterval(state.timer);
    state.timer = 0;
  }
  btnPlay.textContent = "Play";
}

// --- Reset ---

function resetViewerPanels() {
  stepPos.textContent = "Step 0 / 0";
  stepDesc.textContent = "";
  hintEl.classList.remove("hidden");
  explainText.classList.add("hidden");
  bandChip.classList.add("hidden");
  updateTransport();
}

function resetSolveState() {
  stopPlay();
  state.events = [];
  state.baseGrid = null;
  state.step = 0;
  for (const cell of cells) {
    cell.classList.remove(...CELL_CLASSES);
  }
  setStatus("");
  setMetrics("—", "—", "—", "—");
  renderStats(null);
  renderLog();
  resetViewerPanels();
}

loadPuzzles();

// Ensure Go's WebAssembly support is loaded; include wasm_exec.js in index.html before this.
const go = new Go();
let isWasmLoaded = false;

// Load and run the WebAssembly module
async function loadWasm() {
  const resp = await fetch("auction-game.wasm");
  const bytes = await resp.arrayBuffer();
  const result = await WebAssembly.instantiate(bytes, go.importObject);
  go.run(result.instance);
  isWasmLoaded = true;
}

// app.js の先頭あたりに追加
async function performStep() {
  // 1) 現在の State を取得
  let state = window.getCurrentState();
  // 2) 人間待ちなら自動パス、そうでなければ1ステップ実行
  if (state.WaitingHuman) {
    const r = parseInt(document.getElementById("input-r").value, 10) || 0;
    const g = parseInt(document.getElementById("input-g").value, 10) || 0;
    const b = parseInt(document.getElementById("input-b").value, 10) || 0;
    window.submitBid(r, g, b);
  } else {
    await window.nextStep();
  }
  // 3) 更新された State で UI を再描画
  state = window.getCurrentState();
  updateUI(state);
  return state;
}

// Dynamically display player type selectors based on selected number of players
function setupPlayerTypes() {
  const selectPlayers = document.getElementById("select-players");
  const container = document.getElementById("player-types");
  container.innerHTML = "";
  const n = +selectPlayers.value;

  // ★ WASM から AI 名を取得（配列）＋ Human を手動で追加
  const aiNames = window.getAvailableAIs();
  const options = [...aiNames, "Human"];

  for (let i = 0; i < n; i++) {
    const label = document.createElement("label");
    label.textContent = `Player ${i + 1}: `;
    const sel = document.createElement("select");
    sel.id = `type-${i}`;

    options.forEach((name) => {
      const opt = document.createElement("option");
      opt.value = name;
      opt.textContent = name;
      sel.appendChild(opt);
    });
    /* ★ デフォルト選択を決める ------------------------------------ */
    // 例1: プレイヤー0＝Human、その他＝RandomAI
    if (i === 0) {
      sel.value = "Human";
    } else {
      sel.value = "RandomAI"; // ← 好きな AI 名に変更可
    }
    container.appendChild(label);
    container.appendChild(sel);
  }
}

document
  .getElementById("select-players")
  .addEventListener("change", setupPlayerTypes);

function showFinalResults(state) {
  const overlay = document.getElementById("result-overlay");
  const tbody = document.querySelector("#final-table tbody");
  tbody.innerHTML = "";

  // プレイヤーを得点＋コイン合計でソート
  const sorted = [...state.Players].sort((a, b) => {
    if (b.Score !== a.Score) return b.Score - a.Score;
    const sumB = b.Moneys.reduce((s, v) => s + v, 0);
    const sumA = a.Moneys.reduce((s, v) => s + v, 0);
    return sumB - sumA;
  });

  sorted.forEach((p, i) => {
    const tr = document.createElement("tr");
    const coinStr = p.Moneys.join(",");
    const incomeStr = p.Income.join(",");
    [i + 1, p.Index, p.Name, p.Score, coinStr, incomeStr].forEach((val) => {
      const td = document.createElement("td");
      td.textContent = val;
      tr.appendChild(td);
    });
    tbody.appendChild(tr);
  });
  drawScoreChart();
  overlay.style.display = "flex";
}

// 閉じるボタン
document.getElementById("btn-close-results").addEventListener("click", () => {
  document.getElementById("result-overlay").style.display = "none";
});

function updateUI(state) {
  // ゲーム終了時
  if (state.GameOver) {
    // 操作パネルを隠す
    document.getElementById("controls").style.display = "none";
    document.getElementById("human-controls").style.display = "none";
    // 結果オーバーレイを表示
    showFinalResults(state);
    return;
  }
  document.getElementById("phase").textContent = state.Phase;
  document.getElementById("round").textContent = state.Round;

  // 宝石画像の切り替え
  const jewelImg = document.getElementById("jewel-img");
  if (state.Jewel.Point > 0) {
    jewelImg.src = `images/jewel_${state.Jewel.Point}.png`; // 画像は得点に応じたファイル名等に
    jewelImg.style.display = "block";
  } else {
    jewelImg.style.display = "none";
  }
  // Jewel info
  document.getElementById("jewel-point").textContent = state.Jewel.Point;
  document.getElementById("jewel-income").textContent =
    state.Jewel.Income.join(",");

  // Auction info
  document.getElementById("highest-player").textContent =
    state.Auction.MaxPlayer;
  document.getElementById("highest-bid").textContent =
    state.Auction.MaxValue.join(",");

  // Players table
  // Players table
  const tbody = document.getElementById("player-body");
  tbody.innerHTML = "";
  // 手番インデックスを数値に統一
  const turnIndex = Number(state.Turn);
  state.Players.forEach((p) => {
    const pIndex = Number(p.Index);
    const isCurrent = pIndex === turnIndex;
    const tr = document.createElement("tr");
    tr.id = `player-row-${p.Index}`; // ←★ 行に ID を付ける

    // パスしたら淡色化
    if (p.HasPassed) {
      tr.classList.add("passed");
    }
    /* ステータス欄の内容を決定 */
    const statusHtml = p.HasPassed ? "✕" : "";

    /* 最高入札者マーカー列の内容を決定 */
    const isHighest = p.Index === state.Auction.MaxPlayer;
    const bidHtml = isHighest ? '<span class="bid-marker">💰</span>' : "—";

    const cols = [
      p.Index,
      p.Name,
      p.Rank,
      p.Score,
      p.Moneys.join(","),
      p.Income.join(","),
      bidHtml,
      statusHtml,
    ];
    cols.forEach((val, idx) => {
      const td = document.createElement("td");
      // bid／status の末尾2列は innerHTML 挿入
      if (idx >= cols.length - 2) {
        td.innerHTML = val;
      } else {
        td.textContent = val;
      }
      tr.appendChild(td);
    });
    tbody.appendChild(tr);
  });

  document
    .querySelectorAll("#player-body tr.current-turn")
    .forEach((row) => row.classList.remove("current-turn"));
  const hi = document.getElementById(`player-row-${state.Turn}`);
  if (hi) hi.classList.add("current-turn");

  // Human controls
  const humanControls = document.getElementById("human-controls");
  if (state.WaitingHuman) {
    humanControls.style.display = "block";

    // ★ 初期値と min を最高入札額でセット
    const mv = state.Auction.MaxValue; // [r, g, b]
    const inR = document.getElementById("input-r");
    const inG = document.getElementById("input-g");
    const inB = document.getElementById("input-b");

    inR.value = mv[0];
    inG.value = mv[1];
    inB.value = mv[2];

    inR.min = mv[0];
    inG.min = mv[1];
    inB.min = mv[2];
  } else {
    humanControls.style.display = "none";
  }
}

// Bind control buttons to WASM-exported functions
function bindControls() {
  const btnNext = document.getElementById("btn-next");
  const btnSkipR = document.getElementById("btn-skip-round");
  const btnSkipP = document.getElementById("btn-skip-phase");

  // Next ボタン
  btnNext.addEventListener("click", performStep);

  // Skip Round
  btnSkipR.addEventListener("click", async () => {
    let state = await performStep();
    const targetRound = state.Round;
    while (state.Round == targetRound && state.Auction) {
      state = await performStep();
    }
  });

  // Skip Phase
  btnSkipP.addEventListener("click", async () => {
    let state = await performStep();
    const targetPhase = state.Phase;
    while (state.Phase == targetPhase && state.Auction) {
      state = await performStep();
    }
  });

  // ここから追加：キー操作
  document.addEventListener("keydown", (e) => {
    if (!isWasmLoaded) return;
    // ゲーム終了後は auctionState が nil だと nextStep 等が noop になる
    switch (e.key) {
      case "n":
      case "N":
        btnNext.click();
        break;
      case "r":
      case "R":
        btnSkipR.click();
        break;
      case "p":
      case "P":
        btnSkipP.click();
        break;
    }
  });
}

// Bind start button to initialize game
function bindStart() {
  document.getElementById("btn-start").addEventListener("click", () => {
    if (!isWasmLoaded) return;
    const n = parseInt(document.getElementById("select-players").value, 10);
    const types = [];
    for (let i = 0; i < n; i++) {
      types.push(document.getElementById(`type-${i}`).value);
    }
    window.initGame(n, JSON.stringify(types));
    document.getElementById("config").style.display = "none";
    document.getElementById("visualizer").style.display = "block";
    const state = window.getCurrentState();
    updateUI(state);
  });
}

// Bind human bid controls
function bindHumanControls() {
  document.getElementById("btn-human-bid").addEventListener("click", () => {
    const r = parseInt(document.getElementById("input-r").value, 10) || 0;
    const g = parseInt(document.getElementById("input-g").value, 10) || 0;
    const b = parseInt(document.getElementById("input-b").value, 10) || 0;
    window.submitBid(r, g, b);
    const state = window.getCurrentState();
    updateUI(state);
  });
  document.getElementById("btn-human-pass").addEventListener("click", () => {
    window.submitBid(0, 0, 0);
    const state = window.getCurrentState();
    updateUI(state);
  });
}

async function drawScoreChart() {
  const allStates = window.getAllStates();

  // ■ フェーズ開始前スナップショット（PhaseStart===true）を一件ずつ取得
  const phaseMap = new Map(); // phase → state
  allStates.forEach((s) => {
    if (s.PhaseStart && !phaseMap.has(s.Phase)) {
      phaseMap.set(s.Phase, s);
    }
  });
  const phaseStates = Array.from(phaseMap.entries())
    .sort((a, b) => a[0] - b[0])
    .map(([_, state]) => state);

  // ■ 最終結果スナップショットを末尾に追加
  const finalState = allStates.find((s) => s.GameOver);
  if (finalState) phaseStates.push(finalState);

  // ■ ラベル生成：["1","2",…,"10","Final"]
  const labels = phaseStates.map((s) =>
    s.GameOver ? "Final" : s.Phase.toString()
  );
  // ■ 各スナップショット時点の平均点を計算
  //    avgScores[i] は phaseStates[i] の平均
  const avgScores = phaseStates.map((s) => {
    const sum = s.Players.reduce((acc, p) => acc + p.Score, 0);
    return sum / s.Players.length;
  });

  // ■ データセット：各プレイヤーごとに「得点 − 平均」を配列化
  const playerCount = phaseStates[0].Players.length;
  const datasets = Array.from({ length: playerCount }, (_, idx) => ({
    label: `Player ${idx}`,
    data: phaseStates.map((s, i) => s.Players[idx].Score - avgScores[i]),
    fill: false,
  }));

  // ■ Chart.js で描画
  const ctx = document.getElementById("rank-chart").getContext("2d");
  new Chart(ctx, {
    type: "line",
    data: { labels, datasets },
    options: {
      scales: {
        y: {
          reverse: false, // 得点は上向きに増えるので反転不要
          ticks: { beginAtZero: true },
          title: { display: true, text: "得点" },
        },
        x: {
          title: { display: true, text: "フェーズ" },
        },
      },
      plugins: {
        legend: { position: "bottom" },
      },
    },
  });
}

// Initialize visualizer
(async () => {
  await loadWasm();
  setupPlayerTypes();
  bindStart();
  bindControls();
  bindHumanControls();
})();

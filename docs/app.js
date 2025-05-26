let scoreChart = null;

function makeBadges(arr) {
  const classes = ["coin-red", "coin-green", "coin-blue"];
  return arr
    .map((v, i) => `<span class="coin ${classes[i]}">${v}</span>`)
    .join("");
}

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
    label.textContent = `Player ${i}: `;
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
  drawIncomeChart();
  overlay.style.display = "flex";
}

// 閉じるボタン
document.getElementById("btn-close-results").addEventListener("click", () => {
  document.getElementById("result-overlay").style.display = "none";
});

function updateUI(state) {
  const btnShow = document.getElementById("btn-show-results");

  // ゲーム終了時
  if (state.GameOver) {
    // ゲーム終了後は「結果を見る」を表示
    btnShow.style.display = "inline-block";
    // 「入札」部だけ隠す
    document.getElementById("human-controls").style.display = "none";

    // Next／Skip ボタンを無効化（見た目はそのまま）
    document.getElementById("btn-next").disabled = true;
    document.getElementById("btn-skip-round").disabled = true;
    document.getElementById("btn-skip-phase").disabled = true;

    // 結果オーバーレイを表示
    showFinalResults(state);
    return;
  }

  btnShow.style.display = "none";

  // 通常時はボタン有効化（再スタート後にも有効に）
  document.getElementById("btn-next").disabled = false;
  document.getElementById("btn-skip-round").disabled = false;
  document.getElementById("btn-skip-phase").disabled = false;

  document.getElementById("phase").textContent = state.Phase;
  document.getElementById("round").textContent = state.Round;

  // 全フェーズ数・全ラウンド数を計算
  const totalPhases = 10;
  const totalRounds = state.Players.length * 3;

  // 小さな表示とバナー用に fraction 形式でセット
  document.getElementById(
    "phase"
  ).textContent = `${state.Phase}/${totalPhases}`;
  document.getElementById(
    "round"
  ).textContent = `${state.Round}/${totalRounds}`;
  document.getElementById(
    "banner-phase"
  ).textContent = `${state.Phase}/${totalPhases}`;
  document.getElementById(
    "banner-round"
  ).textContent = `${state.Round}/${totalRounds}`;

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
  document.getElementById("highest-bid").innerHTML = makeBadges(
    state.Auction.MaxValue
  );

  // Players table 用に、各コインの最大値を計算
  const maxMoney = state.Players.reduce(
    (acc, p) => p.Moneys.map((v, i) => Math.max(acc[i], v)),
    [0, 0, 0]
  );
  const maxIncomeArr = state.Players.reduce(
    (acc, p) => p.Income.map((v, i) => Math.max(acc[i], v)),
    [0, 0, 0]
  );

  // ヘルパー関数をローカルに定義
  const makeBars = (arr, maxArr) => {
    const colors = ["red", "green", "blue"];
    return (
      `<div class="coin-bar-container">` +
      arr
        .map((v, i) => {
          const pct = maxArr[i] ? (v / maxArr[i]) * 100 : 0;
          return `<div class="coin-bar ${colors[i]}" style="width:${pct}%">
                  <span>${v}</span>
                </div>`;
        })
        .join("") +
      `</div>`
    );
  };

  // Players table の再構築
  const tbody = document.getElementById("player-body");
  tbody.innerHTML = "";
  state.Players.forEach((p) => {
    const tr = document.createElement("tr");
    // ① 行に ID を振る
    tr.id = `player-row-${p.Index}`;

    // ② 現在手番ならハイライト
    if (Number(p.Index) === Number(state.Turn)) {
      tr.classList.add("current-turn");
    }

    if (p.HasPassed) tr.classList.add("passed");
    if (Number(p.Index) === Number(state.Turn))
      tr.classList.add("current-turn");

    // 列データをバー化
    // 基本情報（#, 名前, 順位, 得点）
    [p.Index, p.Name, p.Rank, p.Score].forEach((val) => {
      const td = document.createElement("td");
      td.textContent = val;
      tr.appendChild(td);
    });

    // 所持コイン R,G,B を３セルに分割してバー化
    p.Moneys.forEach((v, i) => {
      const td = document.createElement("td");
      const pct = maxMoney[i] ? (v / maxMoney[i]) * 100 : 0;
      td.innerHTML = `
           <div class="coin-bar-wrapper">
             <span class="coin-bar-label">${v}</span>
             <div class="coin-bar ${
               ["red", "green", "blue"][i]
             }" style="width:${pct}%"></div>
           </div>`;
      tr.appendChild(td);
    });

    // 収入コイン R,G,B を３セルに分割してバー化
    p.Income.forEach((v, i) => {
      const td = document.createElement("td");
      const pct = maxIncomeArr[i] ? (v / maxIncomeArr[i]) * 100 : 0;
      td.innerHTML = `
           <div class="coin-bar-wrapper">
             <span class="coin-bar-label">${v}</span>
             <div class="coin-bar ${
               ["red", "green", "blue"][i]
             }" style="width:${pct}%"></div>
           </div>`;
      tr.appendChild(td);
    });

    // 最高入札マーカー
    const bidTd = document.createElement("td");
    bidTd.innerHTML =
      p.Index === state.Auction.MaxPlayer
        ? '<span class="bid-marker">💰</span>'
        : "—";
    tr.appendChild(bidTd);

    // パスステータス
    const statusTd = document.createElement("td");
    statusTd.textContent = p.HasPassed ? "✕" : "";
    tr.appendChild(statusTd);
    tbody.appendChild(tr);
  });

  document
    .querySelectorAll("#player-body tr.current-turn")
    .forEach((row) => row.classList.remove("current-turn"));
  const hi = document.getElementById(`player-row-${state.Turn}`);
  if (hi) hi.classList.add("current-turn");

  updateScoreLine(state);

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
  const btnShow = document.getElementById("btn-show-results");
  const btnNew = document.getElementById("btn-new-game");

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

  btnShow.addEventListener("click", () => {
    const state = window.getCurrentState();
    showFinalResults(state);
  });

  // 新しいゲーム
  btnNew.addEventListener("click", () => {
    // ビジュアライザを隠して設定画面へ
    document.getElementById("visualizer").style.display = "none";
    document.getElementById("config").style.display = "flex";
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

  if (scoreChart) {
    scoreChart.destroy();
  }

  // ■ Chart.js で描画
  const ctx = document.getElementById("rank-chart").getContext("2d");
  scoreChart = new Chart(ctx, {
    type: "line",
    data: { labels, datasets },
    options: {
      scales: {
        y: {
          reverse: false,
          ticks: { beginAtZero: true },
          title: { display: true, text: "平均からの差分" },
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

async function drawIncomeChart() {
  const allStates = window.getAllStates();

  // フェーズ開始スナップショットを集める（drawScoreChart と同様）
  const phaseMap = new Map();
  allStates.forEach((s) => {
    if (s.PhaseStart && !phaseMap.has(s.Phase)) {
      phaseMap.set(s.Phase, s);
    }
  });
  const phaseStates = Array.from(phaseMap.entries())
    .sort((a, b) => a[0] - b[0])
    .map(([_, state]) => state);

  // 最終結果も末尾に追加
  const finalState = allStates.find((s) => s.GameOver);
  if (finalState) phaseStates.push(finalState);

  // ラベル
  const labels = phaseStates.map((s) =>
    s.GameOver ? "Final" : s.Phase.toString()
  );

  // ■ 各プレイヤーごとのPhaseごとのIncome合計を計算
  const playerCount = phaseStates[0].Players.length;
  const datasets = Array.from({ length: playerCount }, (_, idx) => ({
    label: `Player ${idx}`,
    data: phaseStates.map((s) =>
      s.Players[idx].Income.reduce((sum, c) => sum + c, 0)
    ),
    fill: false,
  }));

  // すでにチャートがあれば破棄
  if (window.incomeChart) {
    window.incomeChart.destroy();
  }

  // 描画
  const ctx2 = document.getElementById("income-chart").getContext("2d");
  window.incomeChart = new Chart(ctx2, {
    type: "line",
    data: {
      labels,
      datasets,
    },
    options: {
      scales: {
        y: {
          beginAtZero: true,
          title: { display: true, text: "Income per Phase" },
        },
        x: {
          title: { display: true, text: "Phase" },
        },
      },
      plugins: {
        legend: { position: "bottom" },
      },
    },
  });
}

function updateScoreLine(state) {
  const container = document.getElementById("score-line-container");

  /* ── ① スコア範囲を決定 ─────────────────────────────── */
  const rawScores = state.Players.map((p) => p.Score);
  const rawMin = Math.min(...rawScores);
  const rawMax = Math.max(...rawScores);
  const rawRange = rawMax - rawMin;

  // 差が 40 未満なら、中央を保ったまま両端を広げて range=40 に
  const mid = (rawMin + rawMax) / 2;
  const displayMin = rawRange < 40 ? mid - 20 : rawMin;
  const displayMax = rawRange < 40 ? mid + 20 : rawMax;
  const range = displayMax - displayMin; // 40 以上を保証
  const padding = 5; // 左右 5 %

  /* ── ② スケール（値→%）関数 ───────────────────────── */
  const toPct = (v) =>
    padding + ((v - displayMin) / range) * (100 - 2 * padding);

  /* ── ③ 一度だけライン要素を用意 ──────────────────── */
  if (!container.querySelector("#score-line")) {
    const line = document.createElement("div");
    line.id = "score-line";
    container.appendChild(line);
  }

  /* ── ④ 旧ラベル・目盛りをクリア ──────────────────── */
  container
    .querySelectorAll(".score-tick, .score-label")
    .forEach((el) => el.remove());

  /* ── ⑤ 両端のラベルと目盛り（実際の rawMin / rawMax） ─ */
  [rawMin, rawMax].forEach((v) => {
    const x = toPct(v) + "%";

    const tick = document.createElement("div");
    tick.className = "score-tick";
    tick.style.left = x;
    container.appendChild(tick);

    const lbl = document.createElement("div");
    lbl.className = "score-label";
    lbl.style.left = x;
    lbl.textContent = v;
    container.appendChild(lbl);
  });

  /* ── ⑥ 同点グループ分け＋色・オフセット設定 ────────── */
  const groups = {};
  state.Players.forEach((p) => (groups[p.Score] ||= []).push(p));

  const colors = [
    "#E53E3E",
    "#2B6CB0",
    "#38A169",
    "#D69E2E",
    "#805AD5",
    "#ED8936",
    "#319795",
    "#D53F8C",
  ];
  const offsetStep = 16; // px・すべて下方向へずらす

  /* ── ⑦ マーカーを配置／更新（差分は CSS transition） ─ */
  state.Players.forEach((p) => {
    const idx = Number(p.Index);
    const grp = groups[p.Score];
    const order = grp.findIndex((x) => x.Index === p.Index); // 同点内順位
    const leftPct = toPct(p.Score);
    const topPx = 50 + order * offsetStep; // 50% 基準で下へ

    let marker = container.querySelector(`#score-marker-${idx}`);
    if (!marker) {
      marker = document.createElement("div");
      marker.id = `score-marker-${idx}`;
      marker.className = "score-marker";
      marker.textContent = idx;
      container.appendChild(marker);
    }
    marker.style.left = leftPct + "%";
    marker.style.top = topPx + "%";
    marker.style.background = colors[idx % colors.length];
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

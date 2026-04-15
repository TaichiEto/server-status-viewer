<script>
  import { onMount, onDestroy } from 'svelte';

  let recentEarthquakes = [];
  let activeEarthquake = null;
  let prefIntensity = {};
  let ws = null;
  let waveKey = 0;
  let epicenterX = 0;
  let epicenterY = 0;
  let connectionStatus = 'connecting';
  let hoveredPref = null;

  // 47都道府県 SVG polygon データ (viewBox 0 0 420 600)
  const prefectures = [
    { id: 1,  name: '北海道',   points: '262,18 392,18 408,58 398,98 358,122 298,138 268,108 258,62' },
    { id: 2,  name: '青森',     points: '252,138 298,138 302,168 272,178 250,165' },
    { id: 3,  name: '岩手',     points: '272,178 302,168 314,208 292,225 268,212' },
    { id: 4,  name: '宮城',     points: '264,212 292,225 298,250 264,258 252,238' },
    { id: 5,  name: '秋田',     points: '230,138 252,138 252,165 235,175 224,163 226,148' },
    { id: 6,  name: '山形',     points: '226,170 252,165 264,212 250,227 230,220 222,196' },
    { id: 7,  name: '福島',     points: '226,227 252,258 256,280 224,280 214,260 216,242' },
    { id: 8,  name: '茨城',     points: '264,258 298,250 312,284 280,297 262,287' },
    { id: 9,  name: '栃木',     points: '250,255 264,258 262,287 242,287 237,270' },
    { id: 10, name: '群馬',     points: '226,250 250,255 240,287 220,284 216,270' },
    { id: 11, name: '埼玉',     points: '240,284 262,284 264,304 244,310 234,300' },
    { id: 12, name: '千葉',     points: '267,284 312,280 322,313 294,330 270,320 264,304' },
    { id: 13, name: '東京',     points: '242,304 264,300 267,320 250,325 240,314' },
    { id: 14, name: '神奈川',   points: '234,317 252,322 257,340 237,347 226,334' },
    { id: 15, name: '新潟',     points: '190,175 226,170 222,217 208,250 190,257 180,238 186,207' },
    { id: 16, name: '富山',     points: '184,257 208,254 212,280 194,284 180,274' },
    { id: 17, name: '石川',     points: '167,250 186,250 186,280 167,287 157,270' },
    { id: 18, name: '福井',     points: '157,280 177,280 177,307 160,310 150,297' },
    { id: 19, name: '山梨',     points: '230,300 244,304 242,324 224,322 220,310' },
    { id: 20, name: '長野',     points: '197,270 230,270 230,300 220,310 197,307 190,287' },
    { id: 21, name: '岐阜',     points: '180,280 204,280 207,314 187,320 172,310 174,294' },
    { id: 22, name: '静岡',     points: '224,324 254,337 257,357 234,367 214,357 210,340' },
    { id: 23, name: '愛知',     points: '194,314 217,314 217,340 197,347 184,334' },
    { id: 24, name: '三重',     points: '187,340 210,334 214,364 200,380 184,370' },
    { id: 25, name: '滋賀',     points: '170,310 190,307 190,334 174,337 167,324' },
    { id: 26, name: '京都',     points: '150,300 174,297 174,324 157,330 144,317' },
    { id: 27, name: '大阪',     points: '157,330 177,327 180,350 164,357 154,347' },
    { id: 28, name: '兵庫',     points: '130,320 157,317 160,350 144,360 124,350 120,334' },
    { id: 29, name: '奈良',     points: '160,344 184,340 187,364 170,374 157,360' },
    { id: 30, name: '和歌山',   points: '157,360 180,354 184,380 167,390 150,377' },
    { id: 31, name: '鳥取',     points: '120,327 150,324 152,344 134,350 117,340' },
    { id: 32, name: '島根',     points: '94,324 123,322 124,344 107,350 90,342' },
    { id: 33, name: '岡山',     points: '134,344 160,340 162,362 144,370 130,360' },
    { id: 34, name: '広島',     points: '107,347 134,344 137,370 120,377 104,370' },
    { id: 35, name: '山口',     points: '80,360 107,354 110,380 92,387 74,380' },
    { id: 36, name: '徳島',     points: '167,374 190,370 194,390 177,397 164,387' },
    { id: 37, name: '香川',     points: '147,362 170,360 172,377 150,382 140,372' },
    { id: 38, name: '愛媛',     points: '120,370 147,367 150,392 134,402 117,392' },
    { id: 39, name: '高知',     points: '130,397 174,392 177,417 150,427 127,417' },
    { id: 40, name: '福岡',     points: '64,364 91,360 94,384 74,392 57,382' },
    { id: 41, name: '佐賀',     points: '50,380 74,377 74,400 57,407 44,397' },
    { id: 42, name: '長崎',     points: '27,374 52,370 57,400 40,414 20,400' },
    { id: 43, name: '熊本',     points: '60,394 90,387 92,417 70,427 54,417' },
    { id: 44, name: '大分',     points: '87,377 110,372 114,400 94,410 80,400' },
    { id: 45, name: '宮崎',     points: '84,410 110,404 114,434 94,447 77,437' },
    { id: 46, name: '鹿児島',   points: '57,424 87,420 90,450 70,464 50,452' },
    { id: 47, name: '沖縄',     points: '40,538 80,533 84,554 64,562 42,552' },
  ];

  // 都道府県名 -> ID マッピング
  const prefNameToId = {
    '北海道': 1, '青森県': 2, '岩手県': 3, '宮城県': 4, '秋田県': 5,
    '山形県': 6, '福島県': 7, '茨城県': 8, '栃木県': 9, '群馬県': 10,
    '埼玉県': 11, '千葉県': 12, '東京都': 13, '神奈川県': 14, '新潟県': 15,
    '富山県': 16, '石川県': 17, '福井県': 18, '山梨県': 19, '長野県': 20,
    '岐阜県': 21, '静岡県': 22, '愛知県': 23, '三重県': 24, '滋賀県': 25,
    '京都府': 26, '大阪府': 27, '兵庫県': 28, '奈良県': 29, '和歌山県': 30,
    '鳥取県': 31, '島根県': 32, '岡山県': 33, '広島県': 34, '山口県': 35,
    '徳島県': 36, '香川県': 37, '愛媛県': 38, '高知県': 39, '福岡県': 40,
    '佐賀県': 41, '長崎県': 42, '熊本県': 43, '大分県': 44, '宮崎県': 45,
    '鹿児島県': 46, '沖縄県': 47,
  };

  const INTENSITY_LEVELS = [
    { scale: 10, label: '1',  color: '#0a3a50', glow: '' },
    { scale: 20, label: '2',  color: '#0a5a78', glow: '' },
    { scale: 30, label: '3',  color: '#0a9070', glow: '' },
    { scale: 40, label: '4',  color: '#e6a817', glow: 'rgba(230,168,23,0.6)' },
    { scale: 45, label: '5弱', color: '#e06020', glow: 'rgba(224,96,32,0.6)' },
    { scale: 50, label: '5強', color: '#cc3010', glow: 'rgba(204,48,16,0.7)' },
    { scale: 55, label: '6弱', color: '#aa0000', glow: 'rgba(170,0,0,0.8)' },
    { scale: 60, label: '6強', color: '#880088', glow: 'rgba(136,0,136,0.8)' },
    { scale: 70, label: '7',  color: '#ff00ff', glow: 'rgba(255,0,255,0.9)' },
  ];

  function scaleToLevel(scale) {
    for (let i = INTENSITY_LEVELS.length - 1; i >= 0; i--) {
      if (scale >= INTENSITY_LEVELS[i - 1]?.scale + 1 || i === 0) {
        if (scale <= INTENSITY_LEVELS[i].scale) {
          return INTENSITY_LEVELS[i];
        }
      }
    }
    return null;
  }

  function scaleToColor(scale) {
    if (scale <= 0) return '#0a1520';
    for (const lv of INTENSITY_LEVELS) {
      if (scale <= lv.scale) return lv.color;
    }
    return '#ff00ff';
  }

  function scaleToLabel(scale) {
    if (!scale || scale <= 0) return '?';
    for (const lv of INTENSITY_LEVELS) {
      if (scale <= lv.scale) return lv.label;
    }
    return '7';
  }

  function scaleToGlow(scale) {
    for (const lv of INTENSITY_LEVELS) {
      if (scale <= lv.scale) return lv.glow;
    }
    return '';
  }

  // 緯度経度 -> SVG座標変換 (日本の範囲: lat 24-46, lon 122-146)
  function geoToSvg(lat, lon) {
    const x = ((lon - 122) / (147 - 122)) * 400 + 10;
    const y = ((46 - lat) / (46 - 24)) * 560 + 20;
    return [x, y];
  }

  function processEarthquake(eq) {
    const newIntensity = {};
    if (eq.points) {
      for (const pt of eq.points) {
        const id = prefNameToId[pt.pref];
        if (id) {
          newIntensity[id] = Math.max(newIntensity[id] || 0, pt.scale);
        }
      }
    }
    prefIntensity = newIntensity;
    activeEarthquake = eq;

    if (eq.earthquake?.hypocenter?.latitude && eq.earthquake?.hypocenter?.longitude) {
      const [x, y] = geoToSvg(
        eq.earthquake.hypocenter.latitude,
        eq.earthquake.hypocenter.longitude
      );
      epicenterX = x;
      epicenterY = y;
      waveKey = Date.now();
    }

    setTimeout(() => {
      prefIntensity = {};
      activeEarthquake = null;
    }, 120000);
  }

  async function fetchRecentEarthquakes() {
    try {
      const res = await fetch('https://api.p2pquake.net/v2/history?codes=551&limit=6');
      if (!res.ok) return;
      const data = await res.json();
      recentEarthquakes = data;
      if (data.length > 0) processEarthquake(data[0]);
    } catch (e) {
      console.error('地震データ取得エラー', e);
    }
  }

  function connectEarthquakeWS() {
    connectionStatus = 'connecting';
    ws = new WebSocket('wss://api.p2pquake.net/v2/ws');
    ws.onopen = () => { connectionStatus = 'connected'; };
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.code === 551) {
          recentEarthquakes = [data, ...recentEarthquakes].slice(0, 6);
          processEarthquake(data);
        }
      } catch (_) {}
    };
    ws.onclose = () => {
      connectionStatus = 'disconnected';
      setTimeout(connectEarthquakeWS, 5000);
    };
    ws.onerror = () => { connectionStatus = 'disconnected'; };
  }

  function formatTime(timeStr) {
    if (!timeStr) return '--';
    // "2024/03/01 12:34:56" -> "03/01 12:34"
    return timeStr.substring(5, 16);
  }

  function magnitudeClass(mag) {
    if (mag >= 7) return 'text-purple-400';
    if (mag >= 6) return 'text-red-400';
    if (mag >= 5) return 'text-orange-400';
    if (mag >= 4) return 'text-yellow-400';
    return 'text-[#00ffcc]';
  }

  onMount(() => {
    fetchRecentEarthquakes();
    connectEarthquakeWS();
  });

  onDestroy(() => {
    ws?.close();
  });
</script>

<div class="border border-red-500/30 bg-[#0d141b] p-4 space-y-4">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <h2 class="text-xs font-bold text-red-400 uppercase tracking-widest flex items-center gap-2">
      <span class="relative flex h-2 w-2">
        <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"></span>
        <span class="relative inline-flex rounded-full h-2 w-2 bg-red-500"></span>
      </span>
      SEISMIC MONITOR // JAPAN
    </h2>
    <div class="flex items-center gap-2 text-[9px]">
      <span class="w-1.5 h-1.5 rounded-full {connectionStatus === 'connected' ? 'bg-green-400' : connectionStatus === 'connecting' ? 'bg-yellow-400 animate-pulse' : 'bg-red-500'}"></span>
      <span class="text-[#64748b] uppercase">{connectionStatus === 'connected' ? 'P2PQUAKE LINK: ACTIVE' : connectionStatus === 'connecting' ? 'LINKING...' : 'LINK LOST'}</span>
    </div>
  </div>

  <div class="flex flex-col xl:flex-row gap-4">
    <!-- SVG Map -->
    <div class="relative flex-shrink-0">
      <svg
        viewBox="0 0 420 600"
        class="w-full xl:w-[260px]"
        style="filter: drop-shadow(0 0 12px rgba(0,40,80,0.8))"
      >
        <defs>
          <pattern id="seismicGrid" width="20" height="20" patternUnits="userSpaceOnUse">
            <path d="M 20 0 L 0 0 0 20" fill="none" stroke="#0d2235" stroke-width="0.5" opacity="0.6"/>
          </pattern>
          <filter id="glow-red">
            <feGaussianBlur stdDeviation="3" result="blur"/>
            <feMerge><feMergeNode in="blur"/><feMergeNode in="SourceGraphic"/></feMerge>
          </filter>
          <filter id="glow-cyan">
            <feGaussianBlur stdDeviation="2" result="blur"/>
            <feMerge><feMergeNode in="blur"/><feMergeNode in="SourceGraphic"/></feMerge>
          </filter>
        </defs>

        <!-- Ocean background -->
        <rect width="420" height="600" fill="#050c14"/>
        <rect width="420" height="600" fill="url(#seismicGrid)"/>

        <!-- Latitude/Longitude reference lines -->
        {#each [130, 135, 140, 145] as lon}
          {@const x = ((lon - 122) / 25) * 400 + 10}
          <line x1={x} y1="0" x2={x} y2="600" stroke="#0d2235" stroke-width="0.8" stroke-dasharray="4,8"/>
          <text x={x + 2} y="598" fill="#0d3050" font-size="6" font-family="monospace">{lon}°E</text>
        {/each}
        {#each [25, 30, 35, 40, 45] as lat}
          {@const y = ((46 - lat) / 22) * 560 + 20}
          <line x1="0" y1={y} x2="420" y2={y} stroke="#0d2235" stroke-width="0.8" stroke-dasharray="4,8"/>
          <text x="2" y={y - 2} fill="#0d3050" font-size="6" font-family="monospace">{lat}°N</text>
        {/each}

        <!-- Prefecture polygons -->
        {#each prefectures as pref}
          {@const scale = prefIntensity[pref.id] || 0}
          {@const fillColor = scaleToColor(scale)}
          {@const hasIntensity = scale > 0}
          {@const glowColor = scaleToGlow(scale)}
          <polygon
            role="img"
            aria-label={pref.name}
            points={pref.points}
            fill={fillColor}
            stroke={hasIntensity ? '#ff5500' : '#00ffcc'}
            stroke-width={hasIntensity ? '1.5' : '0.4'}
            opacity="0.95"
            style={glowColor ? `filter: drop-shadow(0 0 5px ${glowColor})` : ''}
            on:mouseenter={() => hoveredPref = { name: pref.name, scale }}
            on:mouseleave={() => hoveredPref = null}
          >
            <title>{pref.name}{hasIntensity ? ` 震度${scaleToLabel(scale)}` : ''}</title>
          </polygon>
        {/each}

        <!-- Seismic wave animation (SVG SMIL) -->
        {#key waveKey}
          {#if waveKey > 0}
            <g>
              <!-- Outer wave -->
              <circle cx={epicenterX} cy={epicenterY} r="0" fill="none" stroke="#ff2200" stroke-width="2.5">
                <animate attributeName="r" from="0" to="300" dur="4s" fill="freeze" calcMode="ease-out"/>
                <animate attributeName="stroke-opacity" from="0.9" to="0" dur="4s" fill="freeze"/>
              </circle>
              <!-- Mid wave -->
              <circle cx={epicenterX} cy={epicenterY} r="0" fill="none" stroke="#ff6600" stroke-width="1.8">
                <animate attributeName="r" from="0" to="220" dur="3.2s" begin="0.4s" fill="freeze" calcMode="ease-out"/>
                <animate attributeName="stroke-opacity" from="0.8" to="0" dur="3.2s" begin="0.4s" fill="freeze"/>
              </circle>
              <!-- Inner wave -->
              <circle cx={epicenterX} cy={epicenterY} r="0" fill="none" stroke="#ffaa00" stroke-width="1.2">
                <animate attributeName="r" from="0" to="140" dur="2.5s" begin="0.7s" fill="freeze" calcMode="ease-out"/>
                <animate attributeName="stroke-opacity" from="0.7" to="0" dur="2.5s" begin="0.7s" fill="freeze"/>
              </circle>
              <!-- Epicenter marker -->
              <circle cx={epicenterX} cy={epicenterY} r="10" fill="#ff0000" opacity="0.2">
                <animate attributeName="r" from="4" to="14" dur="1s" repeatCount="indefinite" calcMode="ease-out"/>
                <animate attributeName="opacity" from="0.5" to="0" dur="1s" repeatCount="indefinite"/>
              </circle>
              <circle cx={epicenterX} cy={epicenterY} r="4" fill="#ff2200" filter="url(#glow-red)"/>
              <circle cx={epicenterX} cy={epicenterY} r="2" fill="#ffffff"/>
              <!-- Crosshair -->
              <line x1={epicenterX - 10} y1={epicenterY} x2={epicenterX + 10} y2={epicenterY} stroke="#ff4400" stroke-width="0.8" opacity="0.8"/>
              <line x1={epicenterX} y1={epicenterY - 10} x2={epicenterX} y2={epicenterY + 10} stroke="#ff4400" stroke-width="0.8" opacity="0.8"/>
            </g>
          {/if}
        {/key}

        <!-- Hovered prefecture tooltip -->
        {#if hoveredPref}
          <rect x="5" y="5" width="120" height="22" fill="#050c14" stroke="#00ffcc" stroke-width="0.8" opacity="0.9"/>
          <text x="10" y="18" fill="#00ffcc" font-size="9" font-family="monospace">
            {hoveredPref.name}{hoveredPref.scale > 0 ? ` 震度${scaleToLabel(hoveredPref.scale)}` : ' データなし'}
          </text>
        {/if}

        <!-- Map label -->
        <text x="10" y="592" fill="#0d3050" font-size="7" font-family="monospace">P2PQUAKE SEISMIC NET // JMA SCALE</text>
      </svg>
    </div>

    <!-- Right Panel -->
    <div class="flex-1 space-y-4 min-w-0">
      <!-- Active earthquake alert -->
      {#if activeEarthquake?.earthquake}
        {@const eq = activeEarthquake.earthquake}
        {@const mag = eq.hypocenter?.magnitude ?? 0}
        <div class="border border-red-500/60 bg-red-950/20 p-4 relative overflow-hidden">
          <!-- Animated scan line -->
          <div class="absolute inset-0 pointer-events-none overflow-hidden">
            <div class="absolute left-0 right-0 h-px bg-gradient-to-r from-transparent via-red-500/50 to-transparent scan-line"></div>
          </div>
          <div class="text-[9px] text-red-400 uppercase tracking-widest mb-2 flex items-center gap-2">
            <span class="animate-pulse">▶</span> SEISMIC EVENT DETECTED
          </div>
          <div class="flex items-end gap-4 mb-3">
            <div>
              <div class="text-[8px] text-[#64748b] uppercase">MAGNITUDE</div>
              <div class="text-3xl font-black {magnitudeClass(mag)}">M{mag.toFixed(1)}</div>
            </div>
            <div>
              <div class="text-[8px] text-[#64748b] uppercase">MAX INTENSITY</div>
              <div class="text-3xl font-black text-orange-400">
                {eq.maxScale ? '震度' + scaleToLabel(eq.maxScale) : '--'}
              </div>
            </div>
          </div>
          <div class="text-sm font-bold text-white mb-3">{eq.hypocenter?.name ?? '震源不明'}</div>
          <div class="grid grid-cols-3 gap-2 text-[9px]">
            <div class="bg-black/40 p-2 border border-red-500/20">
              <div class="text-[#64748b] uppercase">DEPTH</div>
              <div class="text-red-300 font-bold">{eq.hypocenter?.depth ?? '?'} km</div>
            </div>
            <div class="bg-black/40 p-2 border border-red-500/20">
              <div class="text-[#64748b] uppercase">TIME</div>
              <div class="text-red-300 font-bold text-[8px]">{formatTime(eq.time)}</div>
            </div>
            <div class="bg-black/40 p-2 border border-red-500/20">
              <div class="text-[#64748b] uppercase">TSUNAMI</div>
              <div class="font-bold {eq.domesticTsunami === 'None' ? 'text-green-400' : 'text-red-400 animate-pulse'}">
                {eq.domesticTsunami === 'None' ? 'NONE' : 'ALERT'}
              </div>
            </div>
          </div>
        </div>
      {:else}
        <div class="border border-[#00ffcc]/10 bg-[#050c14] p-4 text-center">
          <div class="text-[10px] text-[#64748b] uppercase">// NO ACTIVE SEISMIC EVENT</div>
          <div class="text-[9px] text-[#00ffcc]/30 mt-1">ALL SENSORS NOMINAL</div>
          <div class="flex justify-center gap-1 mt-3">
            {#each [1,2,3,4,5] as i}
              <div class="w-1 bg-[#00ffcc]/20 rounded-full seismic-idle" style="height: {4 + Math.random() * 12}px; animation-delay: {i * 0.2}s"></div>
            {/each}
          </div>
        </div>
      {/if}

      <!-- Intensity scale legend -->
      <div>
        <div class="text-[8px] text-[#64748b] uppercase mb-2">// JMA INTENSITY SCALE</div>
        <div class="grid grid-cols-4 sm:grid-cols-8 gap-1">
          {#each INTENSITY_LEVELS as lv}
            <div class="flex flex-col items-center gap-1">
              <div class="w-full h-4 border border-[#1e293b]" style="background: {lv.color}; {lv.glow ? `box-shadow: 0 0 6px ${lv.glow}` : ''}"></div>
              <span class="text-[7px] text-[#64748b]">{lv.label}</span>
            </div>
          {/each}
        </div>
      </div>

      <!-- Recent earthquakes -->
      <div>
        <div class="text-[8px] text-[#64748b] uppercase mb-2">// RECENT SEISMIC LOG</div>
        <div class="space-y-1">
          {#each recentEarthquakes.slice(0, 5) as eq, i}
            {@const mag = eq.earthquake?.hypocenter?.magnitude ?? 0}
            <div class="flex items-center gap-2 text-[9px] bg-[#050c14] border border-[#00ffcc]/10 px-2 py-1.5 {i === 0 ? 'border-red-500/30' : ''}">
              <span class="text-[#64748b] text-[7px] w-6 flex-shrink-0">#{i + 1}</span>
              <span class="font-black {magnitudeClass(mag)} w-12 flex-shrink-0">M{mag.toFixed(1)}</span>
              <span class="text-white flex-1 truncate">{eq.earthquake?.hypocenter?.name ?? '不明'}</span>
              <span class="text-[#64748b] text-[7px] flex-shrink-0">{formatTime(eq.earthquake?.time)}</span>
            </div>
          {/each}
          {#if recentEarthquakes.length === 0}
            <div class="text-[9px] text-[#64748b] text-center py-4">データ取得中...</div>
          {/if}
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .scan-line {
    animation: scanDown 3s linear infinite;
    top: 0;
  }

  @keyframes scanDown {
    from { top: 0%; }
    to   { top: 100%; }
  }

  .seismic-idle {
    animation: seismicIdle 1.5s ease-in-out infinite alternate;
  }

  @keyframes seismicIdle {
    from { opacity: 0.2; transform: scaleY(0.5); }
    to   { opacity: 0.6; transform: scaleY(1.3); }
  }
</style>

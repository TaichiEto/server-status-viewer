<script lang="ts">
  import { onMount } from 'svelte';
  import { 
    Users, Cpu, MemoryStick as Memory, Database, Activity, 
    Copy, CheckCircle, HardDrive, ShieldAlert, Monitor, Terminal, 
    Settings, Zap, Clock
  } from 'lucide-svelte';
  import Chart from 'chart.js/auto';

  let stats = {
    system: { hostname: "", os: "", kernel: "", uptime: 0 },
    cpu: { model: "", cores: 0, logical: 0, usage: 0, per_core: [] },
    ram: { total: 0, used: 0, percent: 0 },
    disk: [],
    gpu: [],
    users: [],
    processes: [],
    timestamp: ""
  };

  let filterUser = "";
  let cpuChart;
  let chartInstance;
  let copiedFeedback = "";

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatUptime = (seconds) => {
    const days = Math.floor(seconds / (24 * 3600));
    const hours = Math.floor((seconds % (24 * 3600)) / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    return `${days}d ${hours}h ${mins}m`;
  };

  const showFeedback = (msg) => {
    copiedFeedback = msg;
    setTimeout(() => copiedFeedback = "", 2000);
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    showFeedback(`COMMAND READY: ${text}`);
  };

  const copyKillCommand = (pid) => copyToClipboard(`kill -9 ${pid}`);
  const copyNotifyCommand = (user, tty) => copyToClipboard(`echo "SYSTEM ALERT: Check resource usage." | write ${user} ${tty}`);

  const connectWS = () => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const ws = new WebSocket(`${protocol}//${window.location.host}/ws`);
    ws.onmessage = (event) => {
      stats = JSON.parse(event.data);
      updateCharts();
    };
    ws.onclose = () => setTimeout(connectWS, 3000);
  };

  const updateCharts = () => {
    if (chartInstance) {
      chartInstance.data.labels.push(new Date().toLocaleTimeString());
      chartInstance.data.datasets[0].data.push(stats.cpu.usage);
      if (chartInstance.data.labels.length > 30) {
        chartInstance.data.labels.shift();
        chartInstance.data.datasets[0].data.shift();
      }
      chartInstance.update();
    }
  };

  onMount(() => {
    connectWS();
    chartInstance = new Chart(cpuChart, {
      type: 'line',
      data: {
        labels: [],
        datasets: [{
          label: 'Total Load',
          data: [],
          borderColor: '#00ffcc',
          tension: 0.3,
          fill: true,
          backgroundColor: 'rgba(0, 255, 204, 0.1)',
          borderWidth: 1,
          pointRadius: 0
        }]
      },
      options: { 
        responsive: true, 
        maintainAspectRatio: false, 
        plugins: { legend: { display: false } },
        scales: { 
          x: { display: false },
          y: { min: 0, max: 100, grid: { color: '#1e293b' }, ticks: { color: '#64748b' } } 
        } 
      }
    });
  });

  $: filteredProcesses = filterUser 
    ? stats.processes.filter(p => p.user.includes(filterUser)) 
    : stats.processes.sort((a, b) => b.cpu - a.cpu).slice(0, 50);
</script>

<div class="min-h-screen bg-[#0a0f14] text-[#00ffcc] p-4 font-mono selection:bg-[#00ffcc] selection:text-black">
  <!-- Top Status Bar -->
  <header class="border-2 border-[#00ffcc] p-3 mb-6 bg-[#0d141b] flex flex-wrap justify-between items-center shadow-[0_0_15px_rgba(0,255,204,0.2)]">
    <div class="flex items-center gap-6">
      <div class="flex items-center gap-2">
        <Zap class="animate-pulse" />
        <span class="text-xl font-black tracking-tighter">CORE-CONTROL v4.0</span>
      </div>
      <div class="hidden md:flex gap-4 text-[10px] uppercase text-[#64748b]">
        <div>HOST: <span class="text-[#00ffcc]">{stats.system.hostname}</span></div>
        <div>OS: <span class="text-[#00ffcc]">{stats.system.os}</span></div>
        <div>KERNEL: <span class="text-[#00ffcc]">{stats.system.kernel}</span></div>
      </div>
    </div>
    <div class="flex items-center gap-6">
      <div class="flex items-center gap-2 text-sm">
        <Clock size={16} />
        <span>{formatUptime(stats.system.uptime)}</span>
      </div>
      <div class="text-xs bg-[#00ffcc] text-black px-2 py-1 font-bold">SYSTEM STATUS: NOMINAL</div>
    </div>
  </header>

  <div class="grid grid-cols-1 lg:grid-cols-4 gap-6">
    <!-- CPU & Logical Cores Section -->
    <section class="lg:col-span-2 space-y-6">
      <div class="border border-[#00ffcc]/30 bg-[#0d141b] p-4">
        <div class="flex justify-between items-start mb-4">
          <div>
            <h2 class="text-xs font-bold text-[#64748b] uppercase tracking-widest flex items-center gap-2">
              <Cpu size={14} /> Processor Intelligence
            </h2>
            <div class="text-lg font-bold text-white leading-tight">{stats.cpu.model}</div>
            <div class="text-[10px] text-[#00ffcc]/60 uppercase">{stats.cpu.cores} Physical / {stats.cpu.logical} Logical Cores</div>
          </div>
          <div class="text-4xl font-black">{stats.cpu.usage.toFixed(1)}%</div>
        </div>
        
        <div class="h-32 mb-4 border-b border-[#00ffcc]/10">
          <canvas bind:this={cpuChart}></canvas>
        </div>

        <!-- Logical Cores Grid -->
        <div class="grid grid-cols-4 sm:grid-cols-8 gap-2">
          {#each (stats.cpu.per_core || []) as usage, i}
          <div class="bg-[#161f27] p-1 border border-[#00ffcc]/10">
            <div class="text-[8px] text-[#64748b]">C{i.toString().padStart(2, '0')}</div>
            <div class="h-1 bg-[#1e293b] mt-1 relative overflow-hidden">
              <div class="absolute inset-0 bg-[#00ffcc] transition-all duration-500" style="width: {usage}%"></div>
            </div>
            <div class="text-[9px] text-right mt-0.5 {usage > 80 ? 'text-red-500' : ''}">{usage.toFixed(0)}%</div>
          </div>
          {/each}
        </div>
      </div>

      <!-- Storage Section -->
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        {#each (stats.disk || []) as d}
        <div class="border border-[#00ffcc]/30 bg-[#0d141b] p-4 relative overflow-hidden">
          <div class="absolute top-0 right-0 p-1 opacity-10"><HardDrive size={40} /></div>
          <h3 class="text-[10px] text-[#64748b] uppercase mb-2">Storage: {d.path}</h3>
          <div class="flex justify-between items-end mb-2">
            <div class="text-xl font-bold text-white">{d.percent.toFixed(1)}%</div>
            <div class="text-[10px] text-[#64748b]">{formatBytes(d.used)} / {formatBytes(d.total)}</div>
          </div>
          <div class="h-1 bg-[#1e293b] w-full">
            <div class="h-full bg-cyan-500 shadow-[0_0_8px_#06b6d4]" style="width: {d.percent}%"></div>
          </div>
        </div>
        {/each}
      </div>
    </section>

    <!-- RAM & GPU Section -->
    <section class="lg:col-span-1 space-y-6">
      <div class="border border-purple-500/30 bg-[#0d141b] p-4">
        <h2 class="text-xs font-bold text-purple-400 uppercase tracking-widest flex items-center gap-2 mb-4">
          <Memory size={14} /> Memory Bank
        </h2>
        <div class="text-4xl font-black text-white mb-2">{stats.ram.percent.toFixed(1)}%</div>
        <div class="text-[10px] text-[#64748b] mb-4 uppercase">Used: {formatBytes(stats.ram.used)} | Total: {formatBytes(stats.ram.total)}</div>
        <div class="flex gap-1 h-8">
          {#each Array(20) as _, i}
            <div class="flex-1 {(i/20 * 100) < stats.ram.percent ? 'bg-purple-500 shadow-[0_0_5px_rgba(168,85,247,0.5)]' : 'bg-slate-800'}"></div>
          {/each}
        </div>
      </div>

      {#each (stats.gpu || []) as g}
      <div class="border border-orange-500/30 bg-[#0d141b] p-4 relative">
        <h2 class="text-xs font-bold text-orange-400 uppercase tracking-widest flex items-center gap-2 mb-4">
          <Database size={14} /> Graphics Accelerator
        </h2>
        <div class="text-lg font-bold text-white leading-none mb-1">{g.name}</div>
        <div class="grid grid-cols-2 gap-4 mt-4">
          <div>
            <div class="text-[8px] text-[#64748b] uppercase">Utilization</div>
            <div class="text-xl font-black text-orange-400">{g.util}</div>
          </div>
          <div>
            <div class="text-[8px] text-[#64748b] uppercase">Temperature</div>
            <div class="text-xl font-black text-red-500">{g.temp}</div>
          </div>
        </div>
        <div class="mt-4">
          <div class="text-[8px] text-[#64748b] uppercase mb-1">VRAM Usage: {g.memory_used} / {g.memory_total}</div>
          <div class="h-1 bg-slate-800">
            <div class="h-full bg-orange-500" style="width: {parseFloat(g.util)}%"></div>
          </div>
        </div>
      </div>
      {/each}

      <!-- Users -->
      <div class="border border-emerald-500/30 bg-[#0d141b] p-4">
        <h2 class="text-xs font-bold text-emerald-400 uppercase tracking-widest flex items-center gap-2 mb-4">
          <Users size={14} /> Active Uplinks
        </h2>
        <div class="space-y-2">
          {#each (stats.users || []) as user}
          <div class="flex justify-between items-center text-[10px] bg-emerald-500/5 p-2 border border-emerald-500/10 group">
            <div>
              <div class="font-bold text-emerald-400">@{user.user}</div>
              <div class="text-[#64748b]">{user.source || 'LOCAL'} • {user.terminal}</div>
            </div>
            <button 
              on:click={() => copyNotifyCommand(user.user, user.terminal)}
              class="opacity-0 group-hover:opacity-100 bg-emerald-500/20 hover:bg-emerald-500 text-emerald-400 hover:text-black p-1 transition-all"
            >
              <Copy size={12} />
            </button>
          </div>
          {/each}
        </div>
      </div>
    </section>

    <!-- High-Density Process List -->
    <section class="lg:col-span-1 border border-blue-500/30 bg-[#0d141b] flex flex-col h-[calc(100vh-120px)]">
      <div class="p-4 border-b border-blue-500/20">
        <h2 class="text-xs font-bold text-blue-400 uppercase tracking-widest flex items-center gap-2 mb-4">
          <Activity size={14} /> Process Subsystems
        </h2>
        <input 
          type="text" 
          placeholder="FILTER_BY_USER..." 
          bind:value={filterUser}
          class="w-full bg-black border border-blue-500/30 p-2 text-[10px] focus:outline-none focus:border-blue-500 text-blue-400 uppercase"
        />
      </div>
      <div class="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-blue-500/20">
        <table class="w-full text-[9px]">
          <thead class="sticky top-0 bg-[#0d141b] text-[#64748b] uppercase border-b border-blue-500/20 text-left">
            <tr>
              <th class="p-2">PID</th>
              <th class="p-2">USER</th>
              <th class="p-2">CPU</th>
              <th class="p-2 text-right">ACT</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-blue-500/10">
            {#each (filteredProcesses || []) as proc}
            <tr class="hover:bg-blue-500/5 group transition-colors">
              <td class="p-2 text-blue-500/60">{proc.pid}</td>
              <td class="p-2 font-bold text-white uppercase">{proc.user}</td>
              <td class="p-2 text-blue-400">{proc.cpu.toFixed(1)}%</td>
              <td class="p-2 text-right">
                <button 
                  on:click={() => copyKillCommand(proc.pid)}
                  class="opacity-0 group-hover:opacity-100 bg-red-500/20 hover:bg-red-500 text-red-500 hover:text-black px-2 py-0.5 rounded transition-all"
                >
                  KILL
                </button>
              </td>
            </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </section>
  </div>
</div>

{#if copiedFeedback}
<div class="fixed bottom-8 left-1/2 -translate-x-1/2 bg-[#00ffcc] text-black px-8 py-3 font-black shadow-[0_0_30px_rgba(0,255,204,0.5)] z-50 flex items-center gap-4">
  <ShieldAlert size={20} />
  {copiedFeedback}
</div>
{/if}

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    overflow-x: hidden;
  }
  
  /* CRT Scanline Effect (Optional) */
  :global(body)::after {
    content: " ";
    display: block;
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    right: 0;
    background: linear-gradient(rgba(18, 16, 16, 0) 50%, rgba(0, 0, 0, 0.25) 50%), linear-gradient(90deg, rgba(255, 0, 0, 0.06), rgba(0, 255, 0, 0.02), rgba(0, 0, 255, 0.06));
    z-index: 100;
    background-size: 100% 2px, 3px 100%;
    pointer-events: none;
  }

  .scrollbar-thin::-webkit-scrollbar {
    width: 4px;
  }
  .scrollbar-thin::-webkit-scrollbar-thumb {
    background: #1e293b;
  }
</style>

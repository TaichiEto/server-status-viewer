<script lang="ts">
  import { onMount } from 'svelte';
  import { Users, Cpu, MemoryStick as Memory, Database, Activity, Copy, CheckCircle } from 'lucide-svelte';
  import Chart from 'chart.js/auto';

  let stats = {
    cpu: 0,
    ram: 0,
    gpu: [],
    users: [],
    processes: [],
    timestamp: ""
  };

  let filterUser = "";
  let cpuChart;
  let ramChart;
  let chartInstance;
  let ramChartInstance;
  let copiedFeedback = "";

  const showFeedback = (msg) => {
    copiedFeedback = msg;
    setTimeout(() => copiedFeedback = "", 2000);
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    showFeedback(`Copied: ${text}`);
  };

  const copyKillCommand = (pid) => {
    copyToClipboard(`kill -9 ${pid}`);
  };

  const copyNotifyCommand = (user, tty) => {
    // write command requires message input in terminal
    copyToClipboard(`echo "Hello ${user}, please check your process usage." | write ${user} ${tty}`);
  };

  const connectWS = () => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const ws = new WebSocket(`${protocol}//${window.location.host}/ws`);

    ws.onmessage = (event) => {
      stats = JSON.parse(event.data);
      updateCharts();
    };

    ws.onclose = () => {
      setTimeout(connectWS, 3000);
    };
  };

  const updateCharts = () => {
    if (chartInstance) {
      chartInstance.data.labels.push(new Date().toLocaleTimeString());
      chartInstance.data.datasets[0].data.push(stats.cpu);
      if (chartInstance.data.labels.length > 20) {
        chartInstance.data.labels.shift();
        chartInstance.data.datasets[0].data.shift();
      }
      chartInstance.update();
    }
    if (ramChartInstance) {
      ramChartInstance.data.labels.push(new Date().toLocaleTimeString());
      ramChartInstance.data.datasets[0].data.push(stats.ram);
      if (ramChartInstance.data.labels.length > 20) {
        ramChartInstance.data.labels.shift();
        ramChartInstance.data.datasets[0].data.shift();
      }
      ramChartInstance.update();
    }
  };

  onMount(() => {
    connectWS();
    
    chartInstance = new Chart(cpuChart, {
      type: 'line',
      data: {
        labels: [],
        datasets: [{
          label: 'CPU Usage %',
          data: [],
          borderColor: 'rgb(59, 130, 246)',
          tension: 0.1,
          fill: true,
          backgroundColor: 'rgba(59, 130, 246, 0.1)'
        }]
      },
      options: { responsive: true, maintainAspectRatio: false, scales: { y: { min: 0, max: 100 } } }
    });

    ramChartInstance = new Chart(ramChart, {
      type: 'line',
      data: {
        labels: [],
        datasets: [{
          label: 'RAM Usage %',
          data: [],
          borderColor: 'rgb(168, 85, 247)',
          tension: 0.1,
          fill: true,
          backgroundColor: 'rgba(168, 85, 247, 0.1)'
        }]
      },
      options: { responsive: true, maintainAspectRatio: false, scales: { y: { min: 0, max: 100 } } }
    });
  });

  $: filteredProcesses = filterUser 
    ? stats.processes.filter(p => p.user.includes(filterUser)) 
    : stats.processes;
</script>

<main class="min-h-screen bg-slate-900 text-slate-100 p-4 md:p-8 font-sans">
  <header class="flex justify-between items-center mb-8 border-b border-slate-700 pb-4">
    <div class="flex items-center gap-3">
      <Activity class="text-emerald-400 w-8 h-8" />
      <h1 class="text-2xl font-bold tracking-tight">StatusNode <span class="text-slate-500 font-normal text-sm ml-2">v1.0.0</span></h1>
    </div>
    <div class="text-right">
      <p class="text-slate-400 text-sm">Last Update: {new Date(stats.timestamp).toLocaleTimeString()}</p>
    </div>
  </header>

  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
    <div class="bg-slate-800 rounded-xl p-6 border border-slate-700 shadow-lg">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-slate-400 font-semibold uppercase text-xs tracking-wider">CPU Usage</h2>
        <Cpu class="text-blue-400 w-5 h-5" />
      </div>
      <p class="text-4xl font-bold">{stats.cpu.toFixed(1)}%</p>
      <div class="h-16 mt-4">
        <canvas bind:this={cpuChart}></canvas>
      </div>
    </div>

    <div class="bg-slate-800 rounded-xl p-6 border border-slate-700 shadow-lg">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-slate-400 font-semibold uppercase text-xs tracking-wider">RAM Usage</h2>
        <Memory class="text-purple-400 w-5 h-5" />
      </div>
      <p class="text-4xl font-bold">{stats.ram.toFixed(1)}%</p>
      <div class="h-16 mt-4">
        <canvas bind:this={ramChart}></canvas>
      </div>
    </div>

    {#each stats.gpu as g}
    <div class="bg-slate-800 rounded-xl p-6 border border-slate-700 shadow-lg">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-slate-400 font-semibold uppercase text-xs tracking-wider">GPU: {g.name}</h2>
        <Database class="text-orange-400 w-5 h-5" />
      </div>
      <p class="text-4xl font-bold">{g.util}</p>
      <p class="text-slate-400 text-sm mt-2">{g.memory_used} / {g.memory_total} | {g.temp}</p>
    </div>
    {/each}

    <div class="bg-slate-800 rounded-xl p-6 border border-slate-700 shadow-lg">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-slate-400 font-semibold uppercase text-xs tracking-wider">Active Users</h2>
        <Users class="text-emerald-400 w-5 h-5" />
      </div>
      <p class="text-4xl font-bold">{stats.users.length}</p>
    </div>
  </div>

  <div class="grid grid-cols-1 xl:grid-cols-3 gap-8">
    <!-- Users Table -->
    <div class="xl:col-span-1 bg-slate-800 rounded-xl border border-slate-700 shadow-lg overflow-hidden">
      <div class="p-4 border-b border-slate-700 bg-slate-800/50 flex items-center gap-2">
        <Users class="w-4 h-4 text-emerald-400" />
        <h3 class="font-bold">Active SSH Sessions</h3>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="bg-slate-900/50 text-slate-400 uppercase text-xs">
            <tr>
              <th class="p-3">User</th>
              <th class="p-3">Terminal</th>
              <th class="p-3">Source</th>
              <th class="p-3">Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-700">
            {#each stats.users as user}
            <tr class="hover:bg-slate-700/50 transition-colors">
              <td class="p-3 font-medium text-emerald-400">{user.user}</td>
              <td class="p-3 text-slate-300">{user.terminal}</td>
              <td class="p-3 text-slate-400">{user.source || 'local'}</td>
              <td class="p-3">
                <button 
                  on:click={() => copyNotifyCommand(user.user, user.terminal)}
                  class="bg-slate-700 hover:bg-emerald-600 p-1.5 rounded transition-colors"
                  title="Copy notify command"
                >
                  <Copy class="w-3.5 h-3.5" />
                </button>
              </td>
            </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>

    <!-- Processes Table -->
    <div class="xl:col-span-2 bg-slate-800 rounded-xl border border-slate-700 shadow-lg overflow-hidden">
      <div class="p-4 border-b border-slate-700 bg-slate-800/50 flex justify-between items-center">
        <div class="flex items-center gap-2">
          <Activity class="w-4 h-4 text-blue-400" />
          <h3 class="font-bold">System Processes</h3>
        </div>
        <input 
          type="text" 
          placeholder="Filter by user..." 
          bind:value={filterUser}
          class="bg-slate-900 border border-slate-600 rounded px-3 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-emerald-400 transition-all"
        />
      </div>
      <div class="overflow-x-auto max-h-[600px] overflow-y-auto">
        <table class="w-full text-left text-sm">
          <thead class="bg-slate-900/50 text-slate-400 uppercase text-xs sticky top-0">
            <tr>
              <th class="p-3">PID</th>
              <th class="p-3">User</th>
              <th class="p-3">CPU%</th>
              <th class="p-3">MEM%</th>
              <th class="p-3">Command</th>
              <th class="p-3">Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-700">
            {#each filteredProcesses as proc}
            <tr class="hover:bg-slate-700/50 transition-colors">
              <td class="p-3 text-slate-500 font-mono">{proc.pid}</td>
              <td class="p-3 text-emerald-400 font-medium">{proc.user}</td>
              <td class="p-3 text-blue-400">{proc.cpu.toFixed(1)}%</td>
              <td class="p-3 text-purple-400">{proc.memory.toFixed(1)}%</td>
              <td class="p-3 text-slate-300 truncate max-w-xs" title={proc.command}>{proc.command}</td>
              <td class="p-3">
                <button 
                  on:click={() => copyKillCommand(proc.pid)}
                  class="bg-slate-700 hover:bg-rose-600 p-1.5 rounded transition-colors"
                  title="Copy kill command"
                >
                  <Copy class="w-3.5 h-3.5" />
                </button>
              </td>
            </tr>
            {/each}
          </tbody>

        </table>
      </div>
    </div>
  </div>
</main>

{#if copiedFeedback}
<div class="fixed bottom-8 right-8 bg-emerald-500 text-white px-6 py-3 rounded-lg shadow-2xl flex items-center gap-2 animate-bounce">
  <CheckCircle class="w-5 h-5" />
  <span class="font-bold">{copiedFeedback}</span>
</div>
{/if}

<style>
  :global(body) {
    margin: 0;
    padding: 0;
  }
</style>

const fs = require('fs');
const path = require('path');

// Include scratch node_modules where resvg is installed
const scratchNodeModules = 'C:/Users/alexander/.gemini/antigravity/brain/045b00da-cd6d-47ef-a759-58ba5b70583f/scratch/node_modules';
const { Resvg } = require(path.join(scratchNodeModules, '@resvg/resvg-js'));

function buildLogoSvg() {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="1024" height="1024">
  <defs>
    <clipPath id="squircle-clip">
      <rect x="24" y="24" width="976" height="976" rx="220" />
    </clipPath>

    <linearGradient id="cyan-glow" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#00f5ff"/>
      <stop offset="100%" stop-color="#0284c7"/>
    </linearGradient>

    <linearGradient id="beak-amber" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#fbbf24"/>
      <stop offset="100%" stop-color="#d97706"/>
    </linearGradient>

    <linearGradient id="slate-keel" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#38bdf8"/>
      <stop offset="100%" stop-color="#0284c7"/>
    </linearGradient>

    <filter id="subtle-shadow" x="-10%" y="-10%" width="120%" height="120%">
      <feDropShadow dx="0" dy="12" stdDeviation="16" flood-color="#000000" flood-opacity="0.12" />
    </filter>
  </defs>

  <!-- Luxury White Squircle Container -->
  <rect x="24" y="24" width="976" height="976" rx="220" fill="#ffffff" stroke="#e2e8f0" stroke-width="6" />

  <g clip-path="url(#squircle-clip)">
    <g transform="translate(512, 512)" filter="url(#subtle-shadow)">

      <!-- Hexagonal Architectural Gateway -->
      <polygon points="
        0,-390
        338,-195
        338,195
        0,390
        -338,195
        -338,-195
      " fill="none" stroke="#0f172a" stroke-width="36" stroke-linejoin="round" />

      <polygon points="
        0,-355
        307,-177
        307,177
        0,355
        -307,177
        -307,-177
      " fill="none" stroke="#00f5ff" stroke-width="4" opacity="0.4" stroke-dasharray="16, 12" />

      <!-- ======================================================== -->
      <!-- WATERTIGHT SILHOUETTE (Backing)                          -->
      <!-- ======================================================== -->
      <path d="
        M 0,-295
        L 45,-285 L 65,-250 L 115,-230 L 140,-130 L 255,-35 L 265,110 L 205,215 L 145,160 L 70,265 L 40,315 L 0,330
        L -40,315 L -70,265 L -145,160 L -205,215 L -265,110 L -255,-35 L -140,-130 L -115,-230 L -65,-250 L -45,-285
        Z
      " fill="#0b0f19" />

      <!-- ======================================================== -->
      <!-- FALCON WING VECTOR COVERTS & PRIMARY PLATES              -->
      <!-- ======================================================== -->
      <!-- Left Upper Wing / Shoulder (Light Facet) -->
      <polygon points="-115,-230 -140,-130 -255,-35 -180,30 -90,0" fill="#334155" />
      <polygon points="-140,-130 -255,-35 -170,-45" fill="#475569" />

      <!-- Right Upper Wing / Shoulder (Shadow Facet) -->
      <polygon points="115,-230 140,-130 255,-35 180,30 90,0" fill="#1e293b" />
      <polygon points="140,-130 255,-35 170,-45" fill="#0f172a" />

      <!-- Left Lower Flight Feather Feathers (Swept Vector) -->
      <polygon points="-255,-35 -265,110 -195,200 -180,30" fill="#1e293b" />
      <polygon points="-195,200 -145,160 -90,120 -180,30" fill="#334155" />

      <!-- Right Lower Flight Feathers -->
      <polygon points="255,-35 265,110 195,200 180,30" fill="#0f172a" />
      <polygon points="195,200 145,160 90,120 180,30" fill="#1e293b" />

      <!-- ======================================================== -->
      <!-- CHEST & STERNUM ARMORED PLATES                           -->
      <!-- ======================================================== -->
      <!-- Upper Breastplate -->
      <polygon points="0,-25 -90,0 -85,110 0,75" fill="#334155" />
      <polygon points="0,-25 90,0 85,110 0,75" fill="#1e293b" />

      <!-- Mid Sternum Vectors (Keel) -->
      <polygon points="0,75 -85,110 -65,220 0,185" fill="#475569" />
      <polygon points="0,75 85,110 65,220 0,185" fill="#0f172a" />

      <!-- Lower Keel & Flank Plates -->
      <polygon points="0,185 -65,220 -35,290 0,265" fill="#334155" />
      <polygon points="0,185 65,220 35,290 0,265" fill="#1e293b" />

      <!-- Tail Retrices / Rudder Vector -->
      <polygon points="0,265 -35,290 -30,320 0,330" fill="#1e293b" />
      <polygon points="0,265 35,290 30,320 0,330" fill="#0b0f19" />

      <!-- Vertical Trajectory Guide Inset (Subtle Cyan Needle) -->
      <polygon points="0,85 -6,140 0,175 6,140" fill="url(#cyan-glow)" opacity="0.9" />

      <!-- ======================================================== -->
      <!-- FALCON HEAD, CROWN & SUPRAORBITAL HOOD                   -->
      <!-- ======================================================== -->
      <!-- Central Forehead Crest -->
      <polygon points="0,-295 -45,-285 0,-215" fill="#475569" />
      <polygon points="0,-295 45,-285 0,-215" fill="#1e293b" />

      <!-- Lateral Crown Feathers -->
      <polygon points="-45,-285 -65,-250 -45,-215 0,-215" fill="#334155" />
      <polygon points="45,-285 65,-250 45,-215 0,-215" fill="#0f172a" />

      <!-- Temporal Crests -->
      <polygon points="-65,-250 -115,-230 -80,-175 -45,-215" fill="#1e293b" />
      <polygon points="65,-250 115,-230 80,-175 45,-215" fill="#0b0f19" />

      <!-- Supraorbital Brow Ridge (Key Predatory Feature) -->
      <polygon points="0,-215 -45,-215 -80,-175 -35,-165 0,-175" fill="#475569" />
      <polygon points="0,-215 45,-215 80,-175 35,-165 0,-175" fill="#1e293b" />

      <!-- ======================================================== -->
      <!-- PREDATORY ALMOND EYES (Electric Cyan Optics)             -->
      <!-- ======================================================== -->
      <!-- Left Eye Orbital Socket -->
      <polygon points="-80,-175 -35,-165 -45,-135 -85,-145" fill="#0b0f19" />
      <!-- Left Almond Eye Core -->
      <polygon points="-72,-163 -43,-156 -52,-141 -77,-148" fill="url(#cyan-glow)" />
      <!-- Left Pupil Slit -->
      <polygon points="-62,-160 -54,-157 -56,-144 -64,-147" fill="#032541" />
      <!-- Left Specular Reflection -->
      <circle cx="-60" cy="-155" r="2.5" fill="#ffffff" />

      <!-- Right Eye Orbital Socket -->
      <polygon points="80,-175 35,-165 45,-135 85,-145" fill="#0b0f19" />
      <!-- Right Almond Eye Core -->
      <polygon points="72,-163 43,-156 52,-141 77,-148" fill="url(#cyan-glow)" />
      <!-- Right Pupil Slit -->
      <polygon points="62,-160 54,-157 56,-144 64,-147" fill="#032541" />
      <!-- Right Specular Reflection -->
      <circle cx="58" cy="-155" r="2.5" fill="#ffffff" />

      <!-- ======================================================== -->
      <!-- FALCON MALAR STRIPES (Polygonal Facets, NOT strokes)    -->
      <!-- ======================================================== -->
      <!-- Left Malar Stripe -->
      <polygon points="-85,-145 -45,-135 -65,-75 -95,-95" fill="#0b0f19" />
      <!-- Right Malar Stripe -->
      <polygon points="85,-145 45,-135 65,-75 95,-95" fill="#080b11" />

      <!-- Cheek Highlights -->
      <polygon points="-85,-145 -95,-95 -140,-130 -115,-180" fill="#334155" />
      <polygon points="85,-145 95,-95 140,-130 115,-180" fill="#1e293b" />

      <!-- ======================================================== -->
      <!-- RAPTOR HOOKED BEAK & CERE                               -->
      <!-- ======================================================== -->
      <!-- Cere (Base of Beak) -->
      <polygon points="0,-175 -35,-165 -22,-120 0,-125" fill="url(#beak-amber)" />
      <polygon points="0,-175 35,-165 22,-120 0,-125" fill="#b45309" />

      <!-- Hooked Beak Upper Facets -->
      <polygon points="0,-125 -22,-120 -15,-70 0,-50" fill="#f8fafc" />
      <polygon points="0,-125 22,-120 15,-70 0,-50" fill="#94a3b8" />

      <!-- Sharp Raptor Beak Tip (Hooking Down) -->
      <polygon points="0,-50 -15,-70 0,-30" fill="#cbd5e1" />
      <polygon points="0,-50 15,-70 0,-30" fill="#475569" />

      <!-- Chin & Throat Shield Under Beak -->
      <polygon points="0,-30 -30,-45 -45,-15 0,-5" fill="#1e293b" />
      <polygon points="0,-30 30,-45 45,-15 0,-5" fill="#0f172a" />
      <polygon points="0,-5 -45,-15 -90,0 0,-25" fill="#334155" />
      <polygon points="0,-5 45,-15 90,0 0,-25" fill="#1e293b" />

    </g>
  </g>
</svg>`;
}

async function run() {
  const outputDir = path.resolve('B:/workgit/agentdiff/docs/images');
  const svg = buildLogoSvg();
  const svgPath = path.join(outputDir, 'logo.svg');
  const pngPath = path.join(outputDir, 'logo.png');

  fs.writeFileSync(svgPath, svg, 'utf8');
  console.log('Saved SVG to:', svgPath);

  const resvg = new Resvg(svg, { fitTo: { mode: 'width', value: 1024 } });
  const pngData = resvg.render();
  fs.writeFileSync(pngPath, pngData.asPng());
  console.log('Successfully rendered logo.png at 1024x1024 to:', pngPath);
}

run().catch(err => {
  console.error('Failed to generate logo:', err);
  process.exit(1);
});

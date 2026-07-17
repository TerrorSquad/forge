# forge — brand & graphic asset prompts

Prompts for generating forge's visual assets (Google Stitch, or any image/design
generator). Keep the shared **Brand direction** identical across every asset so
the logo, favicon, and social card read as one system.

## Brand direction (paste at the top of every prompt)

> **Brand:** forge — a git hook runner that runs linters/formatters inside DDEV
> and Docker containers. Developer tool. The name evokes a blacksmith's forge:
> anvil, hammer, molten metal, sparks, heat.
>
> **Aesthetic:** modern developer-tool branding, in the spirit of Vite, Bun,
> Turborepo — bold, minimal, geometric, confident. Flat vector, not skeuomorphic.
> No gradients-as-crutch, no drop shadows, no 3D bevels, no stock-clipart look.
>
> **Palette:**
> - Molten orange/amber `#F97316` (primary, the "heat")
> - Deep amber `#EA580C` (accent)
> - Slate near-black `#0F172A` (dark surfaces / background)
> - Off-white `#F8FAFC` (light surfaces / text on dark)
>
> **Core mark:** a stylized **anvil** (instantly reads "forge", simple enough to
> shrink to a favicon), optionally with 2–3 small spark dots rising from it.
> Avoid words inside the mark.

---

## 1. Logo mark (square, primary)

Use for: site logo, README header, base for the favicon.

> [Brand direction above]
>
> Design a **square logo mark** on a transparent background. A single bold
> geometric **anvil** silhouette in molten orange `#F97316`, with 2–3 small spark
> dots in deeper amber `#EA580C` rising off the top-left horn. Thick, even
> strokes; strong negative space; must stay legible at 32×32 px. No text, no
> letters, no background, no shadow. Flat vector style. Provide on transparent
> and also on a `#0F172A` dark square.
>
> Output: 512×512 PNG (transparent) + SVG if available.

## 2. Favicon

Use for: `website/public/favicon.ico` and browser tab.

> [Brand direction above]
>
> A **radically simplified** version of the forge anvil mark for a 16×16 /
> 32×32 favicon. Just the anvil silhouette, one solid color: molten orange
> `#F97316` anvil on a `#0F172A` rounded-square background. No sparks (too small
> to read), no text, no detail that disappears when tiny. High contrast, chunky.
>
> Output: 512×512 PNG (I'll downscale to .ico at 16/32/48).

## 3. Open Graph social card

Use for: `og:image` / `twitter:image` — shown in Slack, X, Discord, LinkedIn link
previews. **Must be exactly 1200×630 px PNG** (SVG is unreliable for crawlers).

> [Brand direction above]
>
> Design a **1200×630 px social share card**. Dark slate `#0F172A` background
> with a subtle texture of faint spark dots in the lower-right, glowing amber.
> Left side: the word **"forge"** in a heavy geometric sans-serif, off-white
> `#F8FAFC`, lowercase, large. Directly beneath it, one line of tagline in a
> lighter weight, muted slate-grey: **"Run git hooks inside DDEV & Docker."**
> Right side: the orange anvil mark from asset #1, large, with a few bright
> sparks. Generous margins, nothing within 60px of any edge (safe zone). Clean,
> high-contrast, readable as a thumbnail.
>
> Output: exactly 1200×630 PNG.

---

## After generating

Drop the files here (VitePress serves `public/` at the site root under `base`):

```
website/public/logo.svg          # or logo.png — update themeConfig.logo if PNG
website/public/favicon.ico
website/public/og-image.png       # then point og:image/twitter:image at it
```

Then in `website/.vitepress/config.mts`, change the `og:image` / `twitter:image`
`content` from `logo.svg` to `og-image.png`, and add
`['meta', { name: 'twitter:card', content: 'summary_large_image' }]` (swap from
`summary`) so the 1200×630 card renders full-width in previews.

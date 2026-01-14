# Ricambi Manager

Sistema completo di gestione magazzino ricambi auto con interfaccia TUI (Terminal User Interface) realizzato in Go con Bubbletea e MongoDB.

## 🎯 Caratteristiche Principali

### Gestione Articoli
- ✅ Ricerca avanzata (codice, descrizione, barcode, applicabilità)
- ✅ Gestione giacenze e movimenti di magazzino
- ✅ Generazione e scansione codici a barre (EAN-13, Code128)
- ✅ Storico sostituzioni articoli
- ✅ Applicabilità veicoli con ricerca inversa
- ✅ Gestione fornitori per articolo con condizioni commerciali
- ✅ Prezzi netti personalizzati per cliente

### Gestione Clienti
- ✅ Anagrafica completa con categorie
- ✅ Controllo fido in tempo reale con alert
- ✅ Griglie sconti personalizzate e per categoria
- ✅ Budget clienti con obiettivi e scaglioni
- ✅ Buoni a credito con gestione scadenze

### Sistema Commerciale
- ✅ Promozioni con regole di applicabilità
- ✅ Calcolo automatico sconti a cascata
- ✅ Controllo sottocosto/sottoguadagno
- ✅ Kit di vendita con calcolo disponibilità
- ✅ Distinta prezzi netti con scadenza

### Sicurezza e Audit
- ✅ Sistema di autenticazione con hash bcrypt
- ✅ 4 profili predefiniti (Admin, Magazziniere, Venditore, Contabile)
- ✅ Permessi granulari per aree e operazioni
- ✅ Audit log completo delle azioni sensibili
- ✅ Session management con timeout configurabile

### Interfaccia TUI
- ✅ Design professionale con Bubbletea + Lipgloss
- ✅ Navigazione keyboard-driven
- ✅ Tabelle interattive con paginazione
- ✅ Status bar e breadcrumb navigation
- ✅ Alert e notifiche contestuali
- ✅ Help menu dinamico

## 🛠️ Stack Tecnologico

- **Go 1.21+** - Linguaggio di programmazione
- **Bubbletea** - Framework TUI
- **Bubbles** - Componenti UI
- **Lipgloss** - Styling
- **MongoDB 7.0** - Database
- **Docker** - Containerizzazione

## 📋 Requisiti

- Go 1.21 o superiore
- Docker e Docker Compose (per deployment)
- MongoDB 7.0+ (se esecuzione locale)

## 🚀 Quick Start

### 1. Clona il repository

```bash
git clone https://github.com/yourusername/ricambi-manager.git
cd ricambi-manager

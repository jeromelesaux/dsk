# dsk — Bibliothèque Go pour images DSK / HFE / SNA

Ce projet fournit une bibliothèque et un outil en ligne de commande pour manipuler des images de disquettes (DSK/EDSK), des images HFE (GoPlay V4 HxC), des snapshots SNA et des fichiers CPR (cartouches). Le code inclut des utilitaires pour travailler avec les en-têtes AMSDOS et les formats CPC (lecture/écriture, import/export, conversion).

## Fonctionnalités principales

- Lecture et écriture des images DSK (format simple et étendu - EDSK)
- Formatage de disquettes virtuelles (personnalisation: pistes, têtes, secteurs, formats vendor/data)
- Gestion du catalogue AMSDOS : lister, ajouter (`put`), extraire (`get`), afficher en hex/ASCII, supprimer des fichiers
- Import/Export RAW : import/export direct par piste/secteur
- Support complet des en-têtes AMSDOS (création, détection, extraction d'adresses de chargement/exécution)
- Lecture / écriture et conversion HFE ↔ DSK (décodage / encodage MFM)
- Lecture / écriture de snapshots SNA (versions 1, 2, et prise en charge partielle/lecture de la v3)
- Création et manipulation de fichiers CPR (cartouches) : écriture de banques, ajout de fichiers, patch et offsets
- Outils de bas niveau pour lire/écrire blocs et pistes, extraire fichiers bruts et construire des pistes MFM
- CLI complète avec de nombreuses options (formatage, conversion, import/export, analyse, tests automatiques)

## Packages et responsabilités

- `dsk/` : lecture/écriture DSK, catalogue, gestion des blocs, import/export de fichiers AMSDOS.
- `hfe/` : lecture HFE, décodage/encodage MFM, conversion HFE→DSK et DSK→HFE.
- `sna/` : lecture/écriture SNA (v1/v2/v3), import dans SNA, extraction de mémoire.
- `cpr/` : création et manipulation de fichiers CPR (cartouches), gestion de banques.
- `amsdos/` : utils pour détecter et manipuler l'en-tête AMSDOS.
- `cli/` : interface en ligne de commande, exemples d'utilisation.

## Compilation

Compiler l'exécutable CLI :

```bash
go build -o dsk cli/main.go
```

Vous pouvez aussi exécuter directement depuis le répertoire du projet :

```bash
go run ./cli -help
```

## Usage CLI — exemples courants

- Convertir un HFE en DSK :

```bash
./dsk -hfe input.hfe -toDsk output.dsk
```

- Convertir un DSK en HFE :

```bash
./dsk -dsk input.dsk -tohfe output.hfe
```

- Créer une DSK vide (format par défaut) :

```bash
./dsk -dsk new.dsk -format
```

- Lister le contenu d'une DSK :

```bash
./dsk -dsk image.dsk -list
```

- Extraire un fichier AMSDOS depuis une DSK :

```bash
./dsk -dsk image.dsk -get FICHIER.BIN
```

- Insérer un fichier AMSDOS dans une DSK (avec adresse d'exécution/chargement) :

```bash
./dsk -dsk image.dsk -put hello.bin -exec "#1000" -load "#0500"
```

- Créer un SNA et y importer un fichier AMSDOS :

```bash
./dsk -sna out.sna -put hello.bin -snaversion 1 -cpctype 4 -screenmode 1
```

- Import RAW (copie directe à partir d'une piste/secteur) :

```bash
./dsk -dsk image.dsk -put rawfile.bin -rawimport -track 1 -sector 0
```

- Exécuter la suite de tests auto intégrés :

```bash
./dsk -autotest
# ou
go test ./...
```

Pour la liste complète des options, voir le code de l'interface : [cli/main.go](cli/main.go#L1) ou lancer `./dsk -help`.

## Formats supportés

- DSK / EDSK : lecture et écriture, formatage, accès aux pistes et secteurs
- HFE : lecture et écriture (conversion MFM ↔ secteurs)
- SNA : versions 1, 2 (lecture/écriture), v3 partiellement (lecture des chunks CPC+)
- CPR : création et modification de cartouches (32 banques)

## Exemples d'API (Go)

- Lire un DSK depuis le code :

- Voir l'implémentation principale dans [dsk/dsk.go](dsk/dsk.go#L1).

## Tests

Lancer l'ensemble des tests unitaires :

```bash
go test ./...
```

Le binaire CLI contient également une option `-autotest` qui exécute un ensemble d'auto-tests (voir `cli/main.go`).

## Contribuer

Les contributions sont bienvenues : ouvrir des issues pour signaler des bugs ou proposer des améliorations.

## Licence

Voir [LICENSE.md](LICENSE.md#L1) pour les informations de licence.

## Fichiers utiles

- CLI principal : [cli/main.go](cli/main.go#L1)
- Gestion DSK : [dsk/dsk.go](dsk/dsk.go#L1)
- HFE : [hfe/hfe.go](hfe/hfe.go#L1)
- SNA : [sna/sna.go](sna/sna.go#L1)
- CPR : [cpr/cpr.go](cpr/cpr.go#L1)
- AMSDOS helpers : [amsdos/amsdos.go](amsdos/amsdos.go#L1)


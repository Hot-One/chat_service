<template>
  <div class="p-4" @click="closeContextMenu" @contextmenu.prevent="closeContextMenu" @mouseup="handleMouseUp">
    <input type="file" @change="onFileChange" accept=".xlsx" />

    <!-- Отладочная информация -->
    <div v-if="selectedCells.length > 0" class="mt-2 text-sm text-gray-600">
      Выделено ячеек: {{ selectedCells.length }}
    </div>

    <table
      v-if="rows.length"
      class="mt-4 border border-collapse border-gray-300"
    >
      <tr v-for="(row, rowIndex) in rows" :key="rowIndex" class="border-b">
        <td
          v-for="(cell, cellIndex) in row"
          :key="cellIndex"
          class="border px-2 py-1 relative transition-all select-none"
          :class="{
            'bg-blue-100': isCellSelected(rowIndex, cellIndex),
            'ring-2 ring-blue-500 ring-inset z-10': isActiveCell(rowIndex, cellIndex),
            'hover:bg-gray-50': !isCellSelected(rowIndex, cellIndex)
          }"
          :title="cell?.tooltip || ''"
          @contextmenu.prevent="showContextMenu($event, rowIndex, cellIndex)"
          @mousedown="handleMouseDown($event, rowIndex, cellIndex)"
          @mouseenter="handleMouseEnter(rowIndex, cellIndex)"
          @mouseup="handleMouseUp"
        >
          <input
            type="text"
            :value="cell?.display?.result || cell?.display || ''"
            @input="updateCell(rowIndex, cellIndex, $event.target.value)"
            @click.stop
            @keydown="handleKeyDown($event, rowIndex, cellIndex)"
            class="w-full px-1 py-0.5 border-0 outline-none bg-transparent"
            :ref="el => setCellRef(el, rowIndex, cellIndex)"
          />
        </td>
        <td class="border px-2 py-1">
          <button
            @click="deleteRow(rowIndex)"
            class="px-3 py-1 bg-red-500 text-white rounded hover:bg-red-600 text-sm"
          >
            Удалить
          </button>
        </td>
      </tr>
    </table>
    
    <!-- Context Menu -->
    <div
      v-if="contextMenuVisible"
      :style="{
        position: 'fixed',
        left: contextMenuPosition.x + 'px',
        top: contextMenuPosition.y + 'px'
      }"
      class="bg-white border border-gray-300 rounded shadow-lg py-1 z-50"
      @click.stop
    >
      <button
        @click="deleteSelectedCells"
        class="block w-full text-left px-4 py-2 hover:bg-gray-100"
      >
        Удалить {{ selectedCells.length > 1 ? `выделенные ячейки (${selectedCells.length})` : 'ячейку' }}
      </button>
      <button
        @click="clearSelection"
        class="block w-full text-left px-4 py-2 hover:bg-gray-100"
      >
        Очистить выделение
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import ExcelJS from "exceljs";
import * as FormulaParser from "hot-formula-parser";

const rows = ref([]);
const originalData = ref({});
const selectedCells = ref([]);
const activeCell = ref(null);
const anchorCell = ref(null);
const contextMenuVisible = ref(false);
const contextMenuPosition = ref({ x: 0, y: 0 });
const isDragging = ref(false);
const dragStart = ref(null);
const copiedCells = ref([]);

function columnNumberToLetter(column) {
  let letter = "";
  while (column > 0) {
    const mod = (column - 1) % 26;
    letter = String.fromCharCode(65 + mod) + letter;
    column = Math.floor((column - mod) / 26);
  }
  return letter;
}

function createCellContext(worksheet) {
  const context = {};

  worksheet.eachRow((row, rowNumber) => {
    row.eachCell((cell, colNumber) => {
      const colLetter = columnNumberToLetter(colNumber);
      const ref = `${colLetter}${rowNumber}`;
      context[ref] =
        typeof cell?.value === "object" ? cell.value?.result ?? 0 : cell?.value;
    });
  });

  return context;
}

const onFileChange = async (event) => {
  const file = event.target.files[0];
  if (!file) return;

  const buffer = await file.arrayBuffer();
  const workbook = new ExcelJS.Workbook();
  await workbook.xlsx.load(buffer);

  const worksheet = workbook.worksheets[0];
  const extracted = [];

  const parser = new FormulaParser.Parser();
  const cellContext = createCellContext(worksheet);

  parser.on("callCellValue", (cellCoord, done) => {
    const key = `${cellCoord.column}${cellCoord.row}`;
    const value = cellContext[key] ?? 0;
    done(value);
  });

  worksheet.eachRow((row, rowNumber) => {
    const cells = row.values.slice(1).map((cell) => {
      if (cell && typeof cell === "object" && cell.formula) {
        const parsed = parser.parse(cell.formula);
        return {
          display: parsed.error ? "#ERR" : parsed.result,
          tooltip: `=${cell.formula}`,
        };
      }
      return {
        display: cell,
        tooltip: "",
      };
    });

    extracted.push(cells);
  });

  rows.value = extracted;
  originalData.value = { worksheet, workbook, parser, cellContext };
};

const deleteRow = (rowIndex) => {
  rows.value.splice(rowIndex, 1);
  updateCellContext();
};

const deleteCell = (rowIndex, cellIndex) => {
  rows.value[rowIndex][cellIndex] = {
    display: '',
    tooltip: '',
    value: ''
  };
  updateCellContext();
};

const cellRefs = ref({});

const setCellRef = (el, row, col) => {
  if (el) {
    cellRefs.value[`${row}-${col}`] = el;
  }
};

const focusCell = (row, col) => {
  const key = `${row}-${col}`;
  if (cellRefs.value[key]) {
    cellRefs.value[key].focus();
  }
};

const isActiveCell = (row, col) => {
  return activeCell.value && activeCell.value.row === row && activeCell.value.col === col;
};

const handleMouseDown = (event, rowIndex, cellIndex) => {
  event.stopPropagation();
  event.preventDefault();
  
  if (event.button !== 0) return; // Только левая кнопка мыши
  
  if (event.shiftKey) {
    // Выделение диапазона с Shift
    if (anchorCell.value) {
      selectRange(anchorCell.value.row, anchorCell.value.col, rowIndex, cellIndex);
    } else if (selectedCells.value.length > 0) {
      // Если нет anchorCell, но есть выделенная ячейка, используем её
      const firstSelected = selectedCells.value[0];
      selectRange(firstSelected.row, firstSelected.col, rowIndex, cellIndex);
    } else {
      // Если ничего не выделено, просто выделяем эту ячейку
      selectedCells.value = [{ row: rowIndex, col: cellIndex }];
      anchorCell.value = { row: rowIndex, col: cellIndex };
    }
    isDragging.value = false; // Не начинаем drag при Shift
  } else if (event.ctrlKey || event.metaKey) {
    // Добавление/удаление ячеек с Ctrl
    toggleCellSelection(rowIndex, cellIndex);
    isDragging.value = false; // Не начинаем drag при Ctrl
    // Не меняем anchorCell при Ctrl+клик, чтобы сохранить точку для Shift+клик
  } else {
    // Начало нового выделения
    isDragging.value = true;
    dragStart.value = { row: rowIndex, col: cellIndex };
    selectedCells.value = [{ row: rowIndex, col: cellIndex }];
    anchorCell.value = { row: rowIndex, col: cellIndex };
  }
  
  activeCell.value = { row: rowIndex, col: cellIndex };
  focusCell(rowIndex, cellIndex);
};

const handleMouseEnter = (rowIndex, cellIndex) => {
  if (isDragging.value && dragStart.value) {
    selectRange(dragStart.value.row, dragStart.value.col, rowIndex, cellIndex);
  }
};

const handleMouseUp = () => {
  isDragging.value = false;
  dragStart.value = null;
};

const selectRange = (startRow, startCol, endRow, endCol) => {
  const minRow = Math.min(startRow, endRow);
  const maxRow = Math.max(startRow, endRow);
  const minCol = Math.min(startCol, endCol);
  const maxCol = Math.max(startCol, endCol);
  
  console.log('SelectRange:', { startRow, startCol, endRow, endCol, minRow, maxRow, minCol, maxCol });
  
  selectedCells.value = [];
  for (let r = minRow; r <= maxRow; r++) {
    for (let c = minCol; c <= maxCol; c++) {
      if (rows.value[r] && c < rows.value[r].length) {
        selectedCells.value.push({ row: r, col: c });
      }
    }
  }
  console.log('Selected cells count:', selectedCells.value.length);
};

const toggleCellSelection = (row, col) => {
  const index = selectedCells.value.findIndex(
    cell => cell.row === row && cell.col === col
  );
  console.log('Toggle cell:', { row, col, found: index > -1, currentSelection: selectedCells.value.length });
  
  if (index > -1) {
    selectedCells.value.splice(index, 1);
  } else {
    selectedCells.value.push({ row, col });
  }
  console.log('After toggle:', selectedCells.value.length);
};

const handleKeyDown = (event, rowIndex, cellIndex) => {
  const key = event.key;
  const shift = event.shiftKey;
  const ctrl = event.ctrlKey || event.metaKey;
  
  // Навигация стрелками
  if (['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'Tab', 'Enter'].includes(key)) {
    event.preventDefault();
    
    let newRow = rowIndex;
    let newCol = cellIndex;
    
    // Ctrl + стрелки - переход к краю данных
    if (ctrl) {
      switch (key) {
        case 'ArrowUp':
          newRow = 0; // В начало столбца
          break;
        case 'ArrowDown':
          newRow = rows.value.length - 1; // В конец столбца
          break;
        case 'ArrowLeft':
          newCol = 0; // В начало строки
          break;
        case 'ArrowRight':
          newCol = rows.value[rowIndex] ? rows.value[rowIndex].length - 1 : 0; // В конец строки
          break;
      }
    } else {
      // Обычная навигация по одной ячейке
      switch (key) {
        case 'ArrowUp':
          newRow = Math.max(0, rowIndex - 1);
          break;
        case 'ArrowDown':
          newRow = Math.min(rows.value.length - 1, rowIndex + 1);
          break;
        case 'ArrowLeft':
          newCol = Math.max(0, cellIndex - 1);
          break;
        case 'ArrowRight':
          newCol = Math.min(rows.value[rowIndex] ? rows.value[rowIndex].length - 1 : 0, cellIndex + 1);
          break;
        case 'Tab':
          if (shift) {
            newCol = cellIndex > 0 ? cellIndex - 1 : (newRow > 0 ? rows.value[newRow - 1].length - 1 : 0);
            newRow = cellIndex > 0 ? newRow : Math.max(0, newRow - 1);
          } else {
            newCol = cellIndex < (rows.value[rowIndex] ? rows.value[rowIndex].length - 1 : 0) ? cellIndex + 1 : 0;
            newRow = cellIndex < (rows.value[rowIndex] ? rows.value[rowIndex].length - 1 : 0) ? newRow : Math.min(rows.value.length - 1, newRow + 1);
          }
          break;
        case 'Enter':
          newRow = Math.min(rows.value.length - 1, rowIndex + 1);
          break;
      }
    }
    
    if (shift) {
      // Выделение диапазона
      if (!anchorCell.value) {
        anchorCell.value = { row: rowIndex, col: cellIndex };
      }
      console.log('Shift+Arrow: from', anchorCell.value, 'to', { row: newRow, col: newCol });
      selectRange(anchorCell.value.row, anchorCell.value.col, newRow, newCol);
    } else if (!ctrl) {
      // Обычная навигация (без Ctrl)
      selectedCells.value = [{ row: newRow, col: newCol }];
      anchorCell.value = { row: newRow, col: newCol };
    } else {
      // Ctrl + стрелка - переходим но не выделяем
      selectedCells.value = [{ row: newRow, col: newCol }];
      anchorCell.value = { row: newRow, col: newCol };
    }
    
    activeCell.value = { row: newRow, col: newCol };
    focusCell(newRow, newCol);
  }
  
  // Специальные комбинации для выделения
  if (ctrl && shift && ['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight'].includes(key)) {
    event.preventDefault();
    
    if (!anchorCell.value) {
      anchorCell.value = { row: rowIndex, col: cellIndex };
    }
    
    switch (key) {
      case 'ArrowUp':
        // Выделить от текущей позиции до верха столбца
        selectRange(anchorCell.value.row, anchorCell.value.col, 0, cellIndex);
        break;
      case 'ArrowDown':
        // Выделить от текущей позиции до низа столбца
        selectRange(anchorCell.value.row, anchorCell.value.col, rows.value.length - 1, cellIndex);
        break;
      case 'ArrowLeft':
        // Выделить от текущей позиции до начала строки
        selectRange(anchorCell.value.row, anchorCell.value.col, rowIndex, 0);
        break;
      case 'ArrowRight':
        // Выделить от текущей позиции до конца строки
        const maxCol = rows.value[rowIndex] ? rows.value[rowIndex].length - 1 : 0;
        selectRange(anchorCell.value.row, anchorCell.value.col, rowIndex, maxCol);
        break;
    }
    
    activeCell.value = { row: rowIndex, col: cellIndex };
    return;
  }
  
  // Выделить все (Ctrl+A)
  if (ctrl && key === 'a') {
    event.preventDefault();
    selectAllCells();
    return;
  }
  
  // Копирование/вставка
  if (ctrl && key === 'c') {
    event.preventDefault();
    copyCells();
  } else if (ctrl && key === 'v') {
    event.preventDefault();
    pasteCells(rowIndex, cellIndex);
  } else if (key === 'Delete') {
    event.preventDefault();
    deleteSelectedCells();
  }
};

const copyCells = () => {
  copiedCells.value = selectedCells.value.map(({ row, col }) => ({
    row, col,
    value: rows.value[row][col]
  }));
  console.log('Copied', copiedCells.value.length, 'cells');
};

const pasteCells = (targetRow, targetCol) => {
  if (copiedCells.value.length === 0) return;
  
  const minRow = Math.min(...copiedCells.value.map(c => c.row));
  const minCol = Math.min(...copiedCells.value.map(c => c.col));
  
  copiedCells.value.forEach(({ row, col, value }) => {
    const newRow = targetRow + (row - minRow);
    const newCol = targetCol + (col - minCol);
    
    if (rows.value[newRow] && rows.value[newRow][newCol] !== undefined) {
      rows.value[newRow][newCol] = { ...value };
    }
  });
  
  updateCellContext();
};

const selectAllCells = () => {
  selectedCells.value = [];
  
  rows.value.forEach((row, rowIndex) => {
    row.forEach((cell, colIndex) => {
      selectedCells.value.push({ row: rowIndex, col: colIndex });
    });
  });
  
  if (rows.value.length > 0) {
    anchorCell.value = { row: 0, col: 0 };
    activeCell.value = { row: 0, col: 0 };
  }
  
  console.log('Selected all cells:', selectedCells.value.length);
};

const isCellSelected = (rowIndex, cellIndex) => {
  return selectedCells.value.some(
    cell => cell.row === rowIndex && cell.col === cellIndex
  );
};

const showContextMenu = (event, rowIndex, cellIndex) => {
  // Если кликнутая ячейка не выделена, выделить только её
  if (!isCellSelected(rowIndex, cellIndex)) {
    selectedCells.value = [{ row: rowIndex, col: cellIndex }];
  }
  
  contextMenuVisible.value = true;
  contextMenuPosition.value = {
    x: event.clientX,
    y: event.clientY
  };
};

const closeContextMenu = () => {
  contextMenuVisible.value = false;
};

const deleteSelectedCells = () => {
  console.log('Deleting', selectedCells.value.length, 'selected cells');
  
  // Группируем ячейки по строкам и сортируем столбцы в обратном порядке
  const cellsByRow = {};
  selectedCells.value.forEach(({ row, col }) => {
    if (!cellsByRow[row]) cellsByRow[row] = [];
    cellsByRow[row].push(col);
  });
  
  // Удаляем ячейки начиная с конца, чтобы не сбить индексы
  Object.entries(cellsByRow).forEach(([row, cols]) => {
    cols.sort((a, b) => b - a); // Сортируем в обратном порядке
    cols.forEach(col => {
      if (rows.value[row]) {
        rows.value[row].splice(col, 1); // Удаляем ячейку из массива
      }
    });
  });
  
  updateCellContext();
  clearSelection();
  closeContextMenu();
};

const clearSelection = () => {
  selectedCells.value = [];
  activeCell.value = null;
  anchorCell.value = null;
  closeContextMenu();
};

const updateCell = (rowIndex, cellIndex, newValue) => {
  const cell = rows.value[rowIndex][cellIndex];
  
  if (newValue.startsWith('=')) {
    const parser = originalData.value.parser;
    const parsed = parser.parse(newValue.substring(1));
    rows.value[rowIndex][cellIndex] = {
      display: parsed.error ? "#ERR" : parsed.result,
      tooltip: newValue,
      formula: newValue.substring(1)
    };
  } else {
    rows.value[rowIndex][cellIndex] = {
      display: newValue,
      tooltip: "",
      value: newValue
    };
  }
  
  updateCellContext();
};

const updateCellContext = () => {
  if (!originalData.value.parser) return;
  
  const newContext = {};
  rows.value.forEach((row, rowIndex) => {
    row.forEach((cell, colIndex) => {
      const colLetter = columnNumberToLetter(colIndex + 1);
      const ref = `${colLetter}${rowIndex + 1}`;
      newContext[ref] = cell?.display?.result || cell?.display || cell?.value || 0;
    });
  });
  
  originalData.value.cellContext = newContext;
  
  originalData.value.parser.on("callCellValue", (cellCoord, done) => {
    const key = `${cellCoord.column}${cellCoord.row}`;
    const value = newContext[key] ?? 0;
    done(value);
  });
  
  rows.value.forEach((row, rowIndex) => {
    row.forEach((cell, colIndex) => {
      if (cell?.formula) {
        const parsed = originalData.value.parser.parse(cell.formula);
        cell.display = parsed.error ? "#ERR" : parsed.result;
      }
    });
  });
};
</script>

<style scoped>
table {
  border-collapse: collapse;
}
td {
  border: 1px solid #ccc;
}
.ring-2 {
  box-shadow: inset 0 0 0 2px;
}
.ring-blue-500 {
  --tw-ring-color: #3b82f6;
}
.ring-inset {
  box-shadow: inset 0 0 0 2px var(--tw-ring-color);
}
.select-none {
  user-select: none;
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
}
</style>

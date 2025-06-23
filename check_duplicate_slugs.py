#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Script para verificar slugs duplicados en el archivo categories.json
"""

import json
from collections import Counter
from typing import Dict, Set, List

def extract_all_slugs(data: Dict, slugs: Set[str] = None, path: str = "") -> Set[str]:
    """
    Extrae todos los slugs de forma recursiva del JSON de categorías
    
    Args:
        data: Diccionario con los datos de categorías
        slugs: Set para almacenar los slugs encontrados
        path: Ruta actual para debug
    
    Returns:
        Set con todos los slugs encontrados
    """
    if slugs is None:
        slugs = set()
    
    for slug, category_data in data.items():
        # Agregar el slug actual
        slugs.add(slug)
        
        # Si tiene children, procesarlos recursivamente
        if isinstance(category_data, dict) and 'children' in category_data:
            extract_all_slugs(category_data['children'], slugs, f"{path}/{slug}")
    
    return slugs

def find_duplicate_slugs(data: Dict) -> List[str]:
    """
    Encuentra todos los slugs duplicados en el JSON
    
    Args:
        data: Diccionario con los datos de categorías
    
    Returns:
        Lista de slugs duplicados
    """
    # Lista para contar todas las ocurrencias
    all_slugs = []
    
    def collect_slugs(data_dict: Dict, current_path: str = ""):
        """Función interna para recopilar todos los slugs con su ruta"""
        for slug, category_data in data_dict.items():
            all_slugs.append(slug)
            
            # Si tiene children, procesarlos recursivamente
            if isinstance(category_data, dict) and 'children' in category_data:
                collect_slugs(category_data['children'], f"{current_path}/{slug}")
    
    collect_slugs(data)
    
    # Contar ocurrencias
    slug_counts = Counter(all_slugs)
    
    # Encontrar duplicados
    duplicates = [slug for slug, count in slug_counts.items() if count > 1]
    
    return duplicates, slug_counts

def find_slug_locations(data: Dict, target_slug: str, locations: List[str] = None, current_path: str = "") -> List[str]:
    """
    Encuentra todas las ubicaciones donde aparece un slug específico
    
    Args:
        data: Diccionario con los datos de categorías
        target_slug: Slug a buscar
        locations: Lista para almacenar las ubicaciones
        current_path: Ruta actual
    
    Returns:
        Lista de rutas donde aparece el slug
    """
    if locations is None:
        locations = []
    
    for slug, category_data in data.items():
        current_full_path = f"{current_path}/{slug}" if current_path else slug
        
        if slug == target_slug:
            locations.append(current_full_path)
        
        # Si tiene children, procesarlos recursivamente
        if isinstance(category_data, dict) and 'children' in category_data:
            find_slug_locations(category_data['children'], target_slug, locations, current_full_path)
    
    return locations

def main():
    """Función principal"""
    try:
        # Cargar el archivo JSON
        with open('categories.json', 'r', encoding='utf-8') as file:
            categories = json.load(file)
        
        print("🔍 Analizando slugs en categories.json...")
        print("=" * 50)
        
        # Encontrar duplicados
        duplicates, slug_counts = find_duplicate_slugs(categories)
        
        # Mostrar estadísticas generales
        total_slugs = len(slug_counts)
        unique_slugs = len(set(slug_counts.keys()))
        
        print(f"📊 Estadísticas:")
        print(f"   Total de slugs encontrados: {total_slugs}")
        print(f"   Slugs únicos: {unique_slugs}")
        print(f"   Slugs duplicados: {len(duplicates)}")
        print()
        
        if duplicates:
            print("❌ SLUGS DUPLICADOS ENCONTRADOS:")
            print("-" * 30)
            
            for duplicate_slug in duplicates:
                count = slug_counts[duplicate_slug]
                locations = find_slug_locations(categories, duplicate_slug)
                
                print(f"🔴 Slug: '{duplicate_slug}' (aparece {count} veces)")
                print("   Ubicaciones:")
                for location in locations:
                    print(f"     - {location}")
                print()
            
            print("⚠️  ERROR: Se encontraron slugs duplicados. Cada slug debe ser único en toda la jerarquía.")
            return False
        else:
            print("✅ PERFECTO: Todos los slugs son únicos!")
            print("   No se encontraron duplicados en toda la jerarquía.")
            return True
            
    except FileNotFoundError:
        print("❌ Error: No se encontró el archivo 'categories.json'")
        return False
    except json.JSONDecodeError as e:
        print(f"❌ Error al parsear el JSON: {e}")
        return False
    except Exception as e:
        print(f"❌ Error inesperado: {e}")
        return False

if __name__ == "__main__":
    success = main()
    exit(0 if success else 1) 
//MÉTODO QUE SUMA LOS DATOS DE LOS NODOS QUE SE ENCUENTRAN EN LA ALTURA H

float ABB::sumaAux(NodoBin* nodo, unsigned h) const{
    if(nodo == nullptr){
        return 0;
    }

    if(h==1){
        return nodo->dato;
    }
    if(h > 1){
        return sumaAux(nodo->izq, h -1) + sumaAux(nodo->der, h-1);
    }
}

float ABB::suma(unsigned h) const{
    return sumaAux(raiz, h);
}
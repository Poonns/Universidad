// BÚSQUEDA EN PROFUNDIDAD EN UN GRAFO

void Grafo::dfsRec(unsigned origen, vector<bool> & visit){
    visit[origen] = true;
    cout << origen << " ";

    for(int i = 0; i < matriz.size();i++){
        if(matriz[origen][i] > 0 && visit[i] == false){
            dfsRec(i, visit)
        }
    }
}
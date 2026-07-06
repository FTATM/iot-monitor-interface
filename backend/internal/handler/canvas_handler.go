package handler

// import (
// 	"context"
// 	"encoding/json"
// 	"log"
// 	"net/http"
// )

// type CanvasHandler struct {
// 	repo model.UserRepository
// }

// func (app *App) saveLayout(w http.ResponseWriter, r *http.Request) {
// 	var layout []LayoutData

// 	if err := json.NewDecoder(r.Body).Decode(&layout); err != nil {
// 		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
// 		return
// 	}
// 	defer r.Body.Close()

// 	// Convert the Go slice back into raw JSON bytes to store in PostgreSQL JSONB
// 	layoutBytes, err := json.Marshal(layout)
// 	if err != nil {
// 		http.Error(w, "Failed to process layout data", http.StatusInternalServerError)
// 		return
// 	}

// 	// Hardcode a user ID for this example (Normally pulled from a JWT token)
// 	userID := "user_123"

// 	// Raw SQL Upsert: Insert the record. If the user_id exists, update the layout_data instead.
// 	query := `
// 		INSERT INTO dashboard_layouts (user_id, layout_data)
// 		VALUES ($1, $2)
// 		ON CONFLICT (user_id)
// 		DO UPDATE SET
// 			layout_data = EXCLUDED.layout_data,
// 			updated_at = CURRENT_TIMESTAMP;
// 	`

// 	// Execute the query using the connection pool
// 	_, err = app.DB.Exec(context.Background(), query, userID, layoutBytes)
// 	if err != nil {
// 		log.Printf("Database error: %v", err)
// 		http.Error(w, "Failed to save to database", http.StatusInternalServerError)
// 		return
// 	}

// 	// Send success response
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(map[string]string{
// 		"status":  "success",
// 		"message": "Layout saved to PostgreSQL JSONB column!",
// 	})
// }
